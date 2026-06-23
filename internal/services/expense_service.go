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

func (s *ExpenseService) GetGroupExpenses(ctx context.Context, userID string, groupID int64) ([]models.Expense, error) {
	isMember, err := s.groupRepo.IsMember(ctx, groupID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, fmt.Errorf("group not found")
	}

	return s.expenseRepo.GetAllForGroup(ctx, groupID)
}

func (s *ExpenseService) GetExpense(ctx context.Context, userID string, expenseID int64) (*models.Expense, error) {
	expense, err := s.expenseRepo.GetByID(ctx, expenseID)
	if err != nil {
		return nil, err
	}
	if expense == nil {
		return nil, fmt.Errorf("expense not found")
	}

	isMember, err := s.groupRepo.IsMember(ctx, expense.GroupID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, fmt.Errorf("expense not found")
	}

	return expense, nil
}

func (s *ExpenseService) UpdateExpense(ctx context.Context, userID string, expenseID int64, req *models.UpdateExpenseRequest) (*models.Expense, error) {
	expense, err := s.expenseRepo.GetByID(ctx, expenseID)
	if err != nil {
		return nil, err
	}
	if expense == nil {
		return nil, fmt.Errorf("expense not found")
	}

	role, err := s.groupRepo.GetMemberRole(ctx, expense.GroupID, userID)
	if err != nil {
		return nil, err
	}
	if role == "" {
		return nil, fmt.Errorf("expense not found")
	}
	if expense.PaidByUserID != userID && role != "admin" {
		return nil, fmt.Errorf("only the expense creator or a group admin can edit this expense")
	}

	createReq := &models.CreateExpenseRequest{
		Title:        req.Title,
		Amount:       req.Amount,
		Category:     req.Category,
		SplitType:    req.SplitType,
		PaidByUserID: req.PaidByUserID,
		Splits:       req.Splits,
	}

	if err := validateExpenseRequest(createReq); err != nil {
		return nil, err
	}

	calculatedSplits, err := calculateSplits(createReq)
	if err != nil {
		return nil, err
	}

	return s.expenseRepo.Update(ctx, expenseID, req, calculatedSplits)
}
