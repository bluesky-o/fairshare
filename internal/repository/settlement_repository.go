package repository

import (
	"context"
	"fmt"

	"github.com/bluesky-o/fairshare/internal/database"
)

type SettlementRepository struct {
	db *database.DB
}

func NewSettlementRepository(db *database.DB) *SettlementRepository {
	return &SettlementRepository{db: db}
}

func (r *SettlementRepository) GetRawBalances(ctx context.Context, groupID int64) (map[string]float64, error) {
	query := `
		WITH
		paid AS (
			SELECT paid_by_user_id AS user_id, SUM(amount) AS total
			FROM expenses
			WHERE group_id = ?
			GROUP BY paid_by_user_id
		),
		owed AS (
			SELECT es.user_id, SUM(es.owed_amount) AS total
			FROM expense_splits es
			JOIN expenses e ON e.id = es.expense_id
			WHERE e.group_id = ?
			GROUP BY es.user_id
		),
		settlements_paid AS (
			SELECT payer_user_id AS user_id, SUM(amount) AS total
			FROM settlements
			WHERE group_id = ?
			GROUP BY payer_user_id
		),
		settlements_received AS (
			SELECT payee_user_id AS user_id, SUM(amount) AS total
			FROM settlements
			WHERE group_id = ?
			GROUP BY payee_user_id
		),
		all_users AS (
			SELECT user_id FROM group_members WHERE group_id = ?
		)
		SELECT 
			au.user_id,
			COALESCE(p.total, 0)  AS total_paid,
			COALESCE(o.total, 0)  AS total_owed,
			COALESCE(sp.total, 0) AS settlements_paid,
			COALESCE(sr.total, 0) AS settlements_received
		FROM all_users au
		LEFT JOIN paid               p  ON p.user_id  = au.user_id
		LEFT JOIN owed               o  ON o.user_id  = au.user_id
		LEFT JOIN settlements_paid   sp ON sp.user_id = au.user_id
		LEFT JOIN settlements_received sr ON sr.user_id = au.user_id
	`

	rows, err := r.db.QueryContext(ctx, query,
		groupID, groupID, groupID, groupID, groupID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get raw balances: %w", err)
	}
	defer rows.Close()

	balances := make(map[string]float64)

	for rows.Next() {
		var userID string
		var totalPaid, totalOwed, settlementsPaid, settlementsReceived float64

		err := rows.Scan(
			&userID,
			&totalPaid,
			&totalOwed,
			&settlementsPaid,
			&settlementsReceived,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan balance: %w", err)
		}

		net := (totalPaid - totalOwed) + (settlementsPaid - settlementsReceived)
		balances[userID] = net
	}

	return balances, nil
}