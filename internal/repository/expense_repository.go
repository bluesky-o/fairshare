package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/bluesky-o/fairshare/internal/database"
	"github.com/bluesky-o/fairshare/internal/models"
)

type ExpenseRepository struct {
	db *database.DB
}

func NewExpenseRepository(db *database.DB) *ExpenseRepository {
	return &ExpenseRepository{db: db}
}

func (r *ExpenseRepository) GetByID(ctx context.Context, expenseID int64) (*models.Expense, error) {
	expense := &models.Expense{}

	err := r.db.QueryRowContext(ctx, `
		SELECT 
			e.id, e.group_id, e.paid_by_user_id,
			u.display_name,
			e.title, e.amount, e.currency,
			e.category, e.split_type,
			e.expense_date, e.created_at
		FROM expenses e
		JOIN users u ON u.id = e.paid_by_user_id
		WHERE e.id = ?
	`, expenseID).Scan(
		&expense.ID,
		&expense.GroupID,
		&expense.PaidByUserID,
		&expense.PaidByName,
		&expense.Title,
		&expense.Amount,
		&expense.Currency,
		&expense.Category,
		&expense.SplitType,
		&expense.ExpenseDate,
		&expense.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get expense: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT 
			es.id, es.expense_id, es.user_id,
			u.display_name,
			es.owed_amount, es.is_settled, es.settled_at
		FROM expense_splits es
		JOIN users u ON u.id = es.user_id
		WHERE es.expense_id = ?
		ORDER BY es.id ASC
	`, expenseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get splits: %w", err)
	}
	defer rows.Close()

	splits := []models.ExpenseSplit{}
	for rows.Next() {
		var s models.ExpenseSplit
		var isSettledInt int
		err := rows.Scan(
			&s.ID,
			&s.ExpenseID,
			&s.UserID,
			&s.DisplayName,
			&s.OwedAmount,
			&isSettledInt,
			&s.SettledAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan split: %w", err)
		}
		s.IsSettled = isSettledInt == 1
		splits = append(splits, s)
	}
	expense.Splits = splits

	return expense, nil
}

func (r *ExpenseRepository) Create(ctx context.Context, groupID int64, req *models.CreateExpenseRequest, calculatedSplits []models.SplitInput) (*models.Expense, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	expenseDate := time.Now()
	if req.ExpenseDate != nil {
		expenseDate = *req.ExpenseDate
	}

	currency := req.Currency
	if currency == "" {
		currency = "INR"
	}
	category := req.Category
	if category == "" {
		category = "general"
	}

	result, err := tx.ExecContext(ctx, `
		INSERT INTO expenses 
			(group_id, paid_by_user_id, title, amount, currency, category, split_type, expense_date)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`,
		groupID,
		req.PaidByUserID,
		req.Title,
		req.Amount,
		currency,
		category,
		req.SplitType,
		expenseDate,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create expense: %w", err)
	}

	expenseID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get expense id: %w", err)
	}

	for _, split := range calculatedSplits {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO expense_splits (expense_id, user_id, owed_amount)
			VALUES (?, ?, ?)
		`, expenseID, split.UserID, split.OwedAmount)
		if err != nil {
			return nil, fmt.Errorf("failed to create split for user %s: %w", split.UserID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return r.GetByID(ctx, expenseID)
}

func (r *ExpenseRepository) GetAllForGroup(ctx context.Context, groupID int64) ([]models.Expense, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT 
			e.id, e.group_id, e.paid_by_user_id,
			u.display_name,
			e.title, e.amount, e.currency,
			e.category, e.split_type,
			e.expense_date, e.created_at
		FROM expenses e
		JOIN users u ON u.id = e.paid_by_user_id
		WHERE e.group_id = ?
		ORDER BY e.expense_date DESC
	`, groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to get expenses: %w", err)
	}
	defer rows.Close()

	expenses := []models.Expense{}
	for rows.Next() {
		var e models.Expense
		err := rows.Scan(
			&e.ID,
			&e.GroupID,
			&e.PaidByUserID,
			&e.PaidByName,
			&e.Title,
			&e.Amount,
			&e.Currency,
			&e.Category,
			&e.SplitType,
			&e.ExpenseDate,
			&e.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan expense: %w", err)
		}
		expenses = append(expenses, e)
	}

	return expenses, nil
}

func (r *ExpenseRepository) Update(ctx context.Context, expenseID int64, req *models.UpdateExpenseRequest, calculatedSplits []models.SplitInput) (*models.Expense, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
		UPDATE expenses
		SET title = ?, amount = ?, category = ?, 
		    split_type = ?, paid_by_user_id = ?
		WHERE id = ?
	`,
		req.Title,
		req.Amount,
		req.Category,
		req.SplitType,
		req.PaidByUserID,
		expenseID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update expense: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		DELETE FROM expense_splits WHERE expense_id = ?
	`, expenseID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete old splits: %w", err)
	}

	for _, split := range calculatedSplits {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO expense_splits (expense_id, user_id, owed_amount)
			VALUES (?, ?, ?)
		`, expenseID, split.UserID, split.OwedAmount)
		if err != nil {
			return nil, fmt.Errorf("failed to insert split: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit: %w", err)
	}

	return r.GetByID(ctx, expenseID)
}
