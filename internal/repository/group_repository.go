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

func (r *GroupRepository) GetAllForUser(ctx context.Context, userID string) ([]models.Group, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT 
			g.id,
			g.name,
			g.description,
			g.created_by_user_id,
			g.created_at
		FROM groups g
		JOIN group_members gm ON gm.group_id = g.id
		WHERE gm.user_id = ?
		ORDER BY g.created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user groups: %w", err)
	}
	defer rows.Close()

	groups := []models.Group{}
	for rows.Next() {
		var g models.Group
		err := rows.Scan(
			&g.ID,
			&g.Name,
			&g.Description,
			&g.CreatedBy,
			&g.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan group: %w", err)
		}
		groups = append(groups, g)
	}

	return groups, nil
}

func (r *GroupRepository) GetMemberRole(ctx context.Context, groupID int64, userID string) (string, error) {
	var role string
	err := r.db.QueryRowContext(ctx, `
		SELECT role FROM group_members
		WHERE group_id = ? AND user_id = ?
	`, groupID, userID).Scan(&role)

	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to get member role: %w", err)
	}
	return role, nil
}

func (r *GroupRepository) IsMember(ctx context.Context, groupID int64, userID string) (bool, error) {
	role, err := r.GetMemberRole(ctx, groupID, userID)
	if err != nil {
		return false, err
	}
	return role != "", nil
}

func (r *GroupRepository) Update(ctx context.Context, groupID int64, name, description string) (*models.Group, error) {
	_, err := r.db.ExecContext(ctx, `
		UPDATE groups
		SET name = ?, description = ?
		WHERE id = ?
	`, name, description, groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to update group: %w", err)
	}

	return r.GetByID(ctx, groupID, "")
}

func (r *GroupRepository) Delete(ctx context.Context, groupID int64) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM groups WHERE id = ?
	`, groupID)
	if err != nil {
		return fmt.Errorf("failed to delete group: %w", err)
	}
	return nil
}

func (r *GroupRepository) AddMember(ctx context.Context, groupID int64, userID, role string) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO group_members (group_id, user_id, role)
		VALUES (?, ?, ?)
	`, groupID, userID, role)
	if err != nil {
		return fmt.Errorf("failed to add member: %w", err)
	}
	return nil
}

func (r *GroupRepository) RemoveMember(ctx context.Context, groupID int64, userID string) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM group_members
		WHERE group_id = ? AND user_id = ?
	`, groupID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove member: %w", err)
	}
	return nil
}
