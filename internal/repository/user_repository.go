package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/bluesky-o/fairshare/internal/database"
	"github.com/bluesky-o/fairshare/internal/models"
)

type UserRepository struct {
	db *database.DB
}

func NewUserRepository(db *database.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) (*models.User, error) {
	query := `
		INSERT INTO users (id, firebase_uid, email, display_name, avatar_url, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.FirebaseUID,
		user.Email,
		user.DisplayName,
		user.AvatarURL,
		now,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	user.CreatedAt = now
	return user, nil
} 

func (r *UserRepository) GetByFirebaseUID(ctx context.Context, firebaseUID string) (*models.User, error) {
	query := `
		SELECT id, firebase_uid, email, display_name, avatar_url, created_at
		FROM users
		WHERE firebase_uid = ?
		LIMIT 1
	`

	user := &models.User{}

	err := r.db.QueryRowContext(ctx, query, firebaseUID).Scan(
		&user.ID,
		&user.FirebaseUID,
		&user.Email,
		&user.DisplayName,
		&user.AvatarURL,
		&user.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get user by firebase uid: %w", err)
	}

	return user, nil
}
