package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bluesky-o/fairshare/internal/models"
	"github.com/bluesky-o/fairshare/internal/repository"
)

type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) RegisterOrLogin(ctx context.Context, req *models.RegisterRequest) (*models.User, error) {
	existing, err := s.userRepo.GetByFirebaseUID(ctx, req.FirebaseUID)

	if err != nil {
		return nil, fmt.Errorf("failed to find existing user: %w", err)
	}

	if existing != nil {
		return nil, err
	}

	if err := validateRegisterRequest(req); err != nil {
		return nil, err
	}

	user := &models.User{
		ID:          req.FirebaseUID, // Firebase UID IS our user ID
		FirebaseUID: req.FirebaseUID,
		Email:       strings.ToLower(strings.TrimSpace(req.Email)),
		DisplayName: strings.TrimSpace(req.DisplayName),
		AvatarURL:   req.AvatarURL,
		CreatedAt:   time.Now(),
	}

	created, err := s.userRepo.Create(ctx, user)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return created, nil
}

func validateRegisterRequest(req *models.RegisterRequest) error {
	if strings.TrimSpace(req.FirebaseUID) == "" {
		return fmt.Errorf("firebase_uid is required")
	}

	if strings.TrimSpace(req.Email) == "" {
		return fmt.Errorf("email is required")
	}

	if strings.TrimSpace(req.DisplayName) == "" {
		return fmt.Errorf("display_name is required")
	}

	if !strings.Contains(req.Email, "@") {
		return fmt.Errorf("invalid email format")
	}

	return nil
}
