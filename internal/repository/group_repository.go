package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/bluesky-o/fairshare/internal/database"
	"github.com/bluesky-o/fairshare/internal/models"
)

type GroupRepository struct {
	db *database.DB
}

func NewGroupRepository(db *database.DB) *GroupRepository {
	return &GroupRepository{db: db}
}

func (r *GroupRepository) Create(ctx context.Context, name, description, createdByUserID string) (*models.Group, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer tx.Rollback()

	result, err := tx.ExecContext(ctx, `
		INSERT INTO groups (name, description, created_by_user_id)
		VALUES (?, ?, ?)
	`, name, description, createdByUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to create group: %w", err)
	}

	groupID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get group id: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO group_members (group_id, user_id, role)
		VALUES (?, ?, 'admin')
	`, groupID, createdByUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to add creator as member: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return r.GetByID(ctx, groupID, createdByUserID)
}

func (r *GroupRepository) GetByID(ctx context.Context, groupID int64, requestingUserID string) (*models.Group, error) {
	group := &models.Group{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, name, description, created_by_user_id, created_at
		FROM groups
		WHERE id = ?
	`, groupID).Scan(
		&group.ID,
		&group.Name,
		&group.Description,
		&group.CreatedBy,
		&group.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get group: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT 
			gm.user_id,
			u.display_name,
			u.email,
			u.avatar_url,
			gm.role,
			gm.joined_at
		FROM group_members gm
		JOIN users u ON u.id = gm.user_id
		WHERE gm.group_id = ?
		ORDER BY gm.joined_at ASC
	`, groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to get group members: %w", err)
	}
	defer rows.Close()

	members := []models.Member{}
	for rows.Next() {
		var m models.Member
		err := rows.Scan(
			&m.UserID,
			&m.DisplayName,
			&m.Email,
			&m.AvatarURL,
			&m.Role,
			&m.JoinedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan member: %w", err)
		}
		members = append(members, m)
	}
	group.Members = members

	return group, nil
}
