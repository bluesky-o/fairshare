package services

import (
	"context"
	"fmt"
	"math"
	"sort"

	"github.com/bluesky-o/fairshare/internal/models"
	"github.com/bluesky-o/fairshare/internal/repository"
)

type BalanceService struct {
	settlementRepo *repository.SettlementRepository
	groupRepo      *repository.GroupRepository
	userRepo       *repository.UserRepository
	expenseRepo    *repository.ExpenseRepository
}

func NewBalanceService(settlementRepo *repository.SettlementRepository, groupRepo *repository.GroupRepository, userRepo *repository.UserRepository, expenseRepo *repository.ExpenseRepository) *BalanceService {
	return &BalanceService{
		settlementRepo: settlementRepo,
		groupRepo:      groupRepo,
		userRepo:       userRepo,
		expenseRepo:    expenseRepo,
	}
}

func simplifyDebts(balances []models.Balance) []models.DebtSuggestion {
	type entry struct {
		userID string
		name   string
		amount float64
	}

	creditors := []entry{}
	debtors := []entry{}

	for _, b := range balances {
		if b.NetBalance > 0.01 {
			creditors = append(creditors, entry{b.UserID, b.DisplayName, b.NetBalance})
		} else if b.NetBalance < -0.01 {
			debtors = append(debtors, entry{b.UserID, b.DisplayName, -b.NetBalance})
		}
	}

	suggestions := []models.DebtSuggestion{}

	i, j := 0, 0
	for i < len(debtors) && j < len(creditors) {
		debtor := &debtors[i]
		creditor := &creditors[j]

		amount := math.Min(debtor.amount, creditor.amount)
		amount = roundToTwoDecimals(amount)

		if amount > 0.01 {
			suggestions = append(suggestions, models.DebtSuggestion{
				FromUserID:   debtor.userID,
				FromUserName: debtor.name,
				ToUserID:     creditor.userID,
				ToUserName:   creditor.name,
				Amount:       amount,
			})
		}

		debtor.amount -= amount
		creditor.amount -= amount

		if debtor.amount < 0.01 {
			i++
		}
		if creditor.amount < 0.01 {
			j++
		}
	}

	return suggestions
}

func (s *BalanceService) GetGroupBalances(ctx context.Context, userID string, groupID int64) (*models.GroupBalanceResponse, error) {
	isMember, err := s.groupRepo.IsMember(ctx, groupID, userID)
	if err != nil {
		return nil, err
	}

	if !isMember {
		return nil, fmt.Errorf("group not found")
	}

	rawBalances, err := s.settlementRepo.GetRawBalances(ctx, groupID)
	if err != nil {
		return nil, err
	}

	group, err := s.groupRepo.GetByID(ctx, groupID, userID)
	if err != nil {
		return nil, err
	}

	memberMap := make(map[string]models.Member)
	for _, m := range group.Members {
		memberMap[m.UserID] = m
	}

	balances := []models.Balance{}
	for userID, net := range rawBalances {
		member := memberMap[userID]
		balances = append(balances, models.Balance{
			UserID:      userID,
			DisplayName: member.DisplayName,
			AvatarURL:   member.AvatarURL,
			NetBalance:  roundToTwoDecimals(net),
		})
	}

	sort.Slice(balances, func(i, j int) bool {
		return balances[i].NetBalance > balances[j].NetBalance
	})

	suggestions := simplifyDebts(balances)

	expenses, err := s.expenseRepo.GetAllForGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}
	var totalExpenses float64
	for _, e := range expenses {
		totalExpenses += e.Amount
	}

	isSettled := len(suggestions) == 0

	return &models.GroupBalanceResponse{
		Balances:        balances,
		DebtSuggestions: suggestions,
		TotalExpenses:   roundToTwoDecimals(totalExpenses),
		IsSettled:       isSettled,
	}, nil
}
