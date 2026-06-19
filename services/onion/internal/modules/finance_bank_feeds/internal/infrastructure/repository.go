package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
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

func (r *Repository) List(ctx context.Context, businessID int64, limit, offset int32) ([]domain.BankFeedTransaction, int64, error) {
	rows, err := r.queries.ListBankFeedTransactions(ctx, sqlc.ListBankFeedTransactionsParams{
		BusinessID: businessID,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list bank feed transactions",
			logger.Error(err),
			logger.Any("business_id", businessID),
		)
		return nil, 0, err
	}

	items := make([]domain.BankFeedTransaction, 0, len(rows))
	for _, row := range rows {
		items = append(items, domain.BankFeedTransaction{
			ID:               row.ID,
			OwnerID:          row.OwnerID,
			BusinessID:       row.BusinessID,
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

	total, err := r.queries.CountBankFeedTransactions(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to count bank feed transactions",
			logger.Error(err),
			logger.Any("business_id", businessID),
		)
		return nil, 0, err
	}

	return items, total, nil
}

// SumAmountByAccountID returns the net amount (sum of AmountMinorUnits) for
// all feed transactions belonging to a specific bank account. Positive values
// indicate net credits; negative values indicate net debits.
func (r *Repository) SumAmountByAccountID(ctx context.Context, businessID, bankAccountID int64) (int64, error) {
	result, err := r.queries.SumBankFeedAmountByAccount(ctx, sqlc.SumBankFeedAmountByAccountParams{
		BusinessID:    businessID,
		BankAccountID: bankAccountID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to sum bank feed amounts by account", logger.Error(err))
		return 0, err
	}

	return toInt64(result), nil
}

// CountUnmatched returns the number of transactions that have neither a
// matched_expense_id nor a matched_invoice_id.
func (r *Repository) CountUnmatched(ctx context.Context, businessID int64) (int64, error) {
	cnt, err := r.queries.CountUnmatchedBankFeedTransactions(ctx, businessID)
	if err != nil {
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

// toInt64 normalizes the result of COALESCE(SUM(...), 0). MySQL returns the
// sum of a BIGINT column as DECIMAL, which the driver may surface as int64,
// []byte, or string depending on configuration.
func toInt64(v interface{}) int64 {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case int64:
		return val
	case int32:
		return int64(val)
	case float64:
		return int64(val)
	case float32:
		return int64(val)
	case []byte:
		return parseInt64Str(string(val))
	case string:
		return parseInt64Str(val)
	default:
		return parseInt64Str(fmt.Sprintf("%v", val))
	}
}

func parseInt64Str(s string) int64 {
	if s == "" {
		return 0
	}
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		// A DECIMAL sum may carry a fractional/scientific suffix; fall back
		// to float parsing before giving up so the magnitude is preserved.
		if f, ferr := strconv.ParseFloat(s, 64); ferr == nil {
			return int64(f)
		}
		logger.Global().Warn("failed to parse bank feed amount sum from database",
			logger.String("value", s),
			logger.Error(err),
		)
		return 0
	}
	return v
}
