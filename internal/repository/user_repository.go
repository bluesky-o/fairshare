package repository

import "github.com/bluesky-o/fairshare/internal/database"

type UserRepository struct {
	db *database.DB
}

func NewUserRepository(db *database.DB) *UserRepository {
	return &UserRepository{db: db}
}
