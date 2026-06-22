package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/bluesky-o/fairshare/internal/models"
	"github.com/bluesky-o/fairshare/internal/repository"
)

type GroupService struct {
	groupRepo *repository.GroupRepository
	userRepo *repository.UserRepository
}

func NewGroupService(groupRepo *repository.GroupRepository, userRepo *repository.UserRepository) *GroupService {
	return &GroupService{
		groupRepo: groupRepo,
		userRepo: userRepo,
	}
}

func (s *GroupService) CreateGroup(ctx context.Context, userID string, req *models.CreateGroupRequest) (*models.Group, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, fmt.Errorf("group name is required")
	}

	if len(name) > 100 {
		return nil, fmt.Errorf("group name must be under 100 characters")
	}

	return s.groupRepo.Create(ctx, name, strings.TrimSpace(req.Description), userID)
}

func (s *GroupService) GetMyGroups(ctx context.Context, userID string) ([]models.Group, error) {
	return s.groupRepo.GetAllForUser(ctx, userID)
}

func (s *GroupService) GetGroup(ctx context.Context, userID string, groupID int64) (*models.Group, error) {
	group, err := s.groupRepo.GetByID(ctx, groupID, userID)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, fmt.Errorf("group not found")
	}

	isMember, err := s.groupRepo.IsMember(ctx, groupID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, fmt.Errorf("group not found")
	}

	return group, nil
}

func (s *GroupService) UpdateGroup(ctx context.Context, userID string, groupID int64, req *models.UpdateGroupRequest) (*models.Group, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, fmt.Errorf("group name is required")
	}

	role, err := s.groupRepo.GetMemberRole(ctx, groupID, userID)
	if err != nil {
		return nil, err
	}
	if role != "admin" {
		return nil, fmt.Errorf("only admins can update the group")
	}

	return s.groupRepo.Update(ctx, groupID, name, strings.TrimSpace(req.Description))
}

func (s *GroupService) DeleteGroup(ctx context.Context, userID string, groupID int64) error {
	role, err := s.groupRepo.GetMemberRole(ctx, groupID, userID)
	if err != nil {
		return err
	}
	if role != "admin" {
		return fmt.Errorf("only admins can delete the group")
	}

	return s.groupRepo.Delete(ctx, groupID)
}
