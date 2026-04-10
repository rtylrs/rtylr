package infrastructure

import (
	"context"
	"database/sql"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/accounts/internal/domain"
	"github.com/voxire/onion/internal/modules/accounts/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries *sqlc.Queries
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{queries: sqlc.New(dbclient.Pool())}
}

func (r *Repository) List(ctx context.Context, ownerID int64) ([]domain.Account, error) {
	rows, err := r.queries.ListAccounts(ctx, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list accounts", logger.Error(err))
		return nil, err
	}

	items := make([]domain.Account, 0, len(rows))
	for _, row := range rows {
		items = append(items, domain.Account{
			ID:        row.ID,
			OwnerID:   row.OwnerID,
			Name:      row.Name,
			Type:      string(row.Type),
			Currency:  nullStr(row.Currency),
			Balance:   parseFloatNull(row.Balance),
			IsActive:  nullBool(row.IsActive),
			CreatedAt: nullTime(row.CreatedAt),
		})
	}

	return items, nil
}

func (r *Repository) Create(ctx context.Context, ownerID int64, input domain.AccountInput) (int64, error) {
	currency := input.Currency
	if currency == "" {
		currency = "USD"
	}

	res, err := r.queries.CreateAccount(ctx, sqlc.CreateAccountParams{
		OwnerID:  ownerID,
		Name:     input.Name,
		Type:     sqlc.FinanceAccountsType(input.Type),
		Currency: nullableString(currency),
		Balance:  nullableString(fmtFloat(input.Balance)),
		IsActive: sql.NullBool{Bool: input.IsActive, Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create account", logger.Error(err))
		return 0, err
	}

	return res.LastInsertId()
}

// ---------------------------------------------------------------------------
// Conversion helpers
// ---------------------------------------------------------------------------

func nullableString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

func nullStr(s sql.NullString) string {
	if s.Valid {
		return s.String
	}
	return ""
}

func nullBool(b sql.NullBool) bool {
	if b.Valid {
		return b.Bool
	}
	return false
}

func nullTime(t sql.NullTime) time.Time {
	if t.Valid {
		return t.Time
	}
	return time.Time{}
}

func fmtFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', 2, 64)
}

func parseFloatNull(s sql.NullString) float64 {
	if !s.Valid {
		return 0
	}
	v, _ := strconv.ParseFloat(s.String, 64)
	return v
}
