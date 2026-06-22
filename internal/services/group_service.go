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

func (s *GroupService) AddMember(ctx context.Context, requestingUserID string, groupID int64, req *models.AddMemberRequest) (*models.Group, error) {
	isMember, err := s.groupRepo.IsMember(ctx, groupID, requestingUserID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, fmt.Errorf("group not found")
	}

	userToAdd, err := s.userRepo.GetByEmail(ctx, strings.ToLower(strings.TrimSpace(req.Email)))
	if err != nil {
		return nil, err
	}
	if userToAdd == nil {
		return nil, fmt.Errorf("no user found with email %s", req.Email)
	}

	alreadyMember, err := s.groupRepo.IsMember(ctx, groupID, userToAdd.ID)
	if err != nil {
		return nil, err
	}
	if alreadyMember {
		return nil, fmt.Errorf("user is already a member of this group")
	}

	err = s.groupRepo.AddMember(ctx, groupID, userToAdd.ID, "member")
	if err != nil {
		return nil, err
	}

	return s.groupRepo.GetByID(ctx, groupID, requestingUserID)
}

func (s *GroupService) RemoveMember(ctx context.Context, requestingUserID string, groupID int64, targetUserID string) error {
	requestingRole, err := s.groupRepo.GetMemberRole(ctx, groupID, requestingUserID)
	if err != nil {
		return err
	}
	if requestingRole == "" {
		return fmt.Errorf("group not found")
	}

	if requestingRole != "admin" && requestingUserID != targetUserID {
		return fmt.Errorf("you can only remove yourself from the group")
	}

	if targetUserID == requestingUserID && requestingRole == "admin" {
		group, err := s.groupRepo.GetByID(ctx, groupID, requestingUserID)
		if err != nil {
			return err
		}
		adminCount := 0
		for _, m := range group.Members {
			if m.Role == "admin" {
				adminCount++
			}
		}
		if adminCount == 1 {
			return fmt.Errorf("cannot remove the only admin, promote another member first")
		}
	}

	return s.groupRepo.RemoveMember(ctx, groupID, targetUserID)
}
