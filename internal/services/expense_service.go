package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/bluesky-o/fairshare/internal/models"
	"github.com/bluesky-o/fairshare/internal/repository"
)

type ExpenseService struct {
	expenseRepo *repository.ExpenseRepository
	groupRepo   *repository.GroupRepository
}

func NewExpenseService(expenseRepo *repository.ExpenseRepository, groupRepo *repository.GroupRepository) *ExpenseService {
	return &ExpenseService{
		expenseRepo: expenseRepo,
		groupRepo:   groupRepo,
	}
}

func validateExpenseRequest(req *models.CreateExpenseRequest) error {
	if strings.TrimSpace(req.Title) == "" {
		return fmt.Errorf("title is required")
	}
	if req.Amount <= 0 {
		return fmt.Errorf("amount must be greater than 0")
	}
	if req.PaidByUserID == "" {
		return fmt.Errorf("paid_by_user_id is required")
	}
	if len(req.Splits) == 0 {
		return fmt.Errorf("at least one person must be in the split")
	}
	validSplitTypes := map[string]bool{
		"equal": true, "exact": true, "percentage": true,
	}
	if !validSplitTypes[req.SplitType] {
		return fmt.Errorf("split_type must be equal, exact, or percentage")
	}
	return nil
}

func (s *ExpenseService) CreateExpense(ctx context.Context, userID string, groupID int64, req *models.CreateExpenseRequest) (*models.Expense, error) {
	isMember, err := s.groupRepo.IsMember(ctx, groupID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, fmt.Errorf("group not found")
	}

	if err := validateExpenseRequest(req); err != nil {
		return nil, err
	}

	for _, split := range req.Splits {
		isMember, err := s.groupRepo.IsMember(ctx, groupID, split.UserID)
		if err != nil {
			return nil, err
		}
		if !isMember {
			return nil, fmt.Errorf("user %s is not a member of this group", split.UserID)
		}
	}

	payerIsMember, err := s.groupRepo.IsMember(ctx, groupID, req.PaidByUserID)
	if err != nil {
		return nil, err
	}
	if !payerIsMember {
		return nil, fmt.Errorf("payer is not a member of this group")
	}

	calculatedSplits, err := calculateSplits(req)
	if err != nil {
		return nil, err
	}

	return s.expenseRepo.Create(ctx, groupID, req, calculatedSplits)
}
