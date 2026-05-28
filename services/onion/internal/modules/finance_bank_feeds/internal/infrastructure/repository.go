package infrastructure

import (
	"context"
	"database/sql"
	"time"

	"github.com/voxire/onion/internal/modules/finance_bank_feeds/internal/domain"
	"github.com/voxire/onion/internal/modules/finance_bank_feeds/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool    *sql.DB
	queries *sqlc.Queries
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		queries: sqlc.New(db),
		pool:    db,
	}
}

func (r *Repository) List(ctx context.Context, ownerID int64, limit, offset int32) ([]domain.BankFeedTransaction, int64, error) {
	rows, err := r.queries.ListBankFeedTransactions(ctx, sqlc.ListBankFeedTransactionsParams{
		OwnerID: ownerID,
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list bank feed transactions",
			logger.Error(err),
			logger.Any("owner_id", ownerID),
		)
		return nil, 0, err
	}

	items := make([]domain.BankFeedTransaction, 0, len(rows))
	for _, row := range rows {
		items = append(items, domain.BankFeedTransaction{
			ID:               row.ID,
			OwnerID:          row.OwnerID,
			BankAccountID:    row.BankAccountID,
			AmountMinorUnits: row.AmountMinorUnits,
			Currency:         row.Currency,
			OccurredAt:       row.OccurredAt,
			Description:      nullStr(row.Description),
			ExternalID:       row.ExternalID,
			MatchedExpenseID: nullInt64Ptr(row.MatchedExpenseID),
			MatchedInvoiceID: nullInt64Ptr(row.MatchedInvoiceID),
			CreatedAt:        nullTime(row.CreatedAt),
		})
	}

	total, err := r.queries.CountBankFeedTransactions(ctx, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to count bank feed transactions",
			logger.Error(err),
			logger.Any("owner_id", ownerID),
		)
		return nil, 0, err
	}

	return items, total, nil
}

// SumAmountByAccountID returns the net amount (sum of AmountMinorUnits) for
// all feed transactions belonging to a specific bank account. Positive values
// indicate net credits; negative values indicate net debits.
func (r *Repository) SumAmountByAccountID(ctx context.Context, ownerID, bankAccountID int64) (int64, error) {
	const q = `
		SELECT COALESCE(SUM(amount_minor_units), 0)
		FROM finance_bank_feed_transactions
		WHERE owner_id = ? AND bank_account_id = ?`

	var sum int64
	if err := r.pool.QueryRowContext(ctx, q, ownerID, bankAccountID).Scan(&sum); err != nil {
		logger.Ctx(ctx).Error("Failed to sum bank feed amounts by account", logger.Error(err))
		return 0, err
	}

	return sum, nil
}

// CountUnmatched returns the number of transactions that have neither a
// matched_expense_id nor a matched_invoice_id.
func (r *Repository) CountUnmatched(ctx context.Context, ownerID int64) (int64, error) {
	const q = `
		SELECT COUNT(*)
		FROM finance_bank_feed_transactions
		WHERE owner_id = ?
		  AND matched_expense_id IS NULL
		  AND matched_invoice_id IS NULL`

	var cnt int64
	if err := r.pool.QueryRowContext(ctx, q, ownerID).Scan(&cnt); err != nil {
		logger.Ctx(ctx).Error("Failed to count unmatched bank feed transactions", logger.Error(err))
		return 0, err
	}

	return cnt, nil
}

// ---------------------------------------------------------------------------
// Conversion helpers
// ---------------------------------------------------------------------------

func nullStr(s sql.NullString) string {
	if s.Valid {
		return s.String
	}
	return ""
}

func nullInt64Ptr(n sql.NullInt64) *int64 {
	if !n.Valid {
		return nil
	}
	v := n.Int64
	return &v
}

func nullTime(t sql.NullTime) time.Time {
	if t.Valid {
		return t.Time
	}
	return time.Time{}
}
