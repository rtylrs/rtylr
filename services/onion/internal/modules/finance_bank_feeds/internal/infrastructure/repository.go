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
