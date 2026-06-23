package models

import "time"

type ExpenseSplit struct {
	ID          int64     `json:"id"`
	ExpenseID   int64     `json:"expense_id"`
	UserID      string    `json:"user_id"`
	DisplayName string    `json:"display_name"`
	OwedAmount  float64   `json:"owed_amount"`
	IsSettled   bool      `json:"is_settled"`
	SettledAt   *time.Time `json:"settled_at,omitempty"`
}

type CreateExpenseRequest struct {
	Title       string       `json:"title"`
	Amount      float64      `json:"amount"`
	Currency    string       `json:"currency,omitempty"`
	Category    string       `json:"category,omitempty"`
	SplitType   string       `json:"split_type"` // "equal", "exact", "percentage"
	PaidByUserID string      `json:"paid_by_user_id"`
	Splits      []SplitInput `json:"splits"`
	ExpenseDate *time.Time   `json:"expense_date,omitempty"`
}

type SplitInput struct {
	UserID     string  `json:"user_id"`
	OwedAmount float64 `json:"owed_amount,omitempty"`
	Percentage float64 `json:"percentage,omitempty"`
}

type Expense struct {
	ID            int64         `json:"id"`
	GroupID       int64         `json:"group_id"`
	PaidByUserID  string        `json:"paid_by_user_id"`
	PaidByName    string        `json:"paid_by_name"`
	Title         string        `json:"title"`
	Amount        float64       `json:"amount"`
	Currency      string        `json:"currency"`
	Category      string        `json:"category"`
	SplitType     string        `json:"split_type"`
	Splits        []ExpenseSplit `json:"splits,omitempty"`
	ExpenseDate   time.Time     `json:"expense_date"`
	CreatedAt     time.Time     `json:"created_at"`
}
