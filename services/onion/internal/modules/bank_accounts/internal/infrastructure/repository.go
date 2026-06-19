package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/bank_accounts/internal/domain"
	"github.com/voxire/onion/internal/modules/bank_accounts/internal/infrastructure/persistence/sqlc"
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

func (r *Repository) List(ctx context.Context, businessID int64) ([]domain.BankAccount, error) {
	rows, err := r.queries.ListBankAccounts(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list bank accounts", logger.Error(err))
		return nil, err
	}

	items := make([]domain.BankAccount, 0, len(rows))
	for _, row := range rows {
		items = append(items, bankAccountFromSqlc(row))
	}

	return items, nil
}

func (r *Repository) Create(ctx context.Context, businessID int64, input domain.BankAccountInput) (int64, error) {
	currency := input.Currency
	if currency == "" {
		currency = "USD"
	}

	res, err := r.queries.CreateBankAccount(ctx, sqlc.CreateBankAccountParams{
		OwnerID:     businessID,
		BusinessID:  businessID,
		Name:        input.Name,
		Institution: nullableString(input.Institution),
		AccountType: nullableString(input.AccountType),
		Balance:     nullableString(fmtFloat(input.Balance)),
		Currency:    nullableString(currency),
		IsConnected: sql.NullBool{Bool: input.IsConnected, Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create bank account", logger.Error(err))
		return 0, err
	}

	return res.LastInsertId()
}

// TotalBalance returns the sum of all account balances for the owner.
func (r *Repository) TotalBalance(ctx context.Context, businessID int64) (float64, error) {
	result, err := r.queries.SumBankAccountBalances(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to sum bank account balances", logger.Error(err))
		return 0, err
	}

	return toFloat64(result), nil
}

// CountConnected returns the number of bank accounts with is_connected = TRUE.
func (r *Repository) CountConnected(ctx context.Context, businessID int64) (int64, error) {
	cnt, err := r.queries.CountConnectedBankAccounts(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to count connected bank accounts", logger.Error(err))
		return 0, err
	}

	return cnt, nil
}

// ---------------------------------------------------------------------------
// Mapping helpers
// ---------------------------------------------------------------------------

func bankAccountFromSqlc(row sqlc.FinanceBankAccount) domain.BankAccount {
	ba := domain.BankAccount{
		ID:          row.ID,
		OwnerID:     row.OwnerID,
		BusinessID:  row.BusinessID,
		Name:        row.Name,
		Institution: nullStr(row.Institution),
		AccountType: nullStr(row.AccountType),
		Balance:     parseFloatNull(row.Balance),
		Currency:    nullStr(row.Currency),
		IsConnected: nullBool(row.IsConnected),
		CreatedAt:   nullTime(row.CreatedAt),
	}
	if row.LastSyncedAt.Valid {
		ba.LastSyncedAt = row.LastSyncedAt.Time
	}
	return ba
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
	v, err := strconv.ParseFloat(s.String, 64)
	if err != nil {
		logger.Global().Warn("failed to parse bank account balance from database",
			logger.String("value", s.String),
			logger.Error(err),
		)
		return 0
	}
	return v
}

// toFloat64 converts the value returned by a COALESCE(SUM(...), 0) aggregate
// (sqlc emits interface{} for DECIMAL aggregates) into a float64, handling the
// driver representations MySQL may return for a DECIMAL column.
func toFloat64(v interface{}) float64 {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int64:
		return float64(val)
	case []byte:
		return parseFloatStr(string(val))
	case string:
		return parseFloatStr(val)
	default:
		return parseFloatStr(fmt.Sprintf("%v", val))
	}
}

func parseFloatStr(s string) float64 {
	if s == "" {
		return 0
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		logger.Global().Warn("failed to parse bank account balance sum from database",
			logger.String("value", s),
			logger.Error(err),
		)
		return 0
	}
	return v
}
