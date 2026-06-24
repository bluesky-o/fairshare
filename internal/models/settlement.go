package models

type GroupBalanceResponse struct {
	Balances        []Balance        `json:"balances"`
	DebtSuggestions []DebtSuggestion `json:"debt_suggestions"`
	TotalExpenses   float64          `json:"total_expenses"`
	IsSettled       bool             `json:"is_settled"`
}

type Balance struct {
	UserID      string  `json:"user_id"`
	DisplayName string  `json:"display_name"`
	AvatarURL   string  `json:"avatar_url,omitempty"`
	NetBalance  float64 `json:"net_balance"`
}

type DebtSuggestion struct {
	FromUserID   string  `json:"from_user_id"`
	FromUserName string  `json:"from_user_name"`
	ToUserID     string  `json:"to_user_id"`
	ToUserName   string  `json:"to_user_name"`
	Amount       float64 `json:"amount"`
}
