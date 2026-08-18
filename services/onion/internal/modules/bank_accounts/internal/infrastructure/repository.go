package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/bank_accounts/internal/domain"
	"github.com/voxire/onion/internal/modules/bank_accounts/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/money"
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
		Balance:     nullableString(input.Balance.String()),
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
func (r *Repository) TotalBalance(ctx context.Context, businessID int64) (money.Amount, error) {
	result, err := r.queries.SumBankAccountBalances(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to sum bank account balances", logger.Error(err))
		return 0, err
	}

	return toMoney(result), nil
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
		Balance:     parseMoneyNull(row.Balance),
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

func parseMoneyNull(s sql.NullString) money.Amount {
	if !s.Valid {
		return 0
	}
	value, err := money.FromString(s.String)
	if err != nil {
		logger.Global().Warn("failed to parse bank account balance from database",
			logger.String("value", s.String),
			logger.Error(err),
		)
		return 0
	}
	return value
}

func toMoney(v interface{}) money.Amount {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case money.Amount:
		return val
	case float64:
		parsed, err := money.FromFloat64(val)
		if err == nil {
			return parsed
		}
	case float32:
		parsed, err := money.FromFloat64(float64(val))
		if err == nil {
			return parsed
		}
	case int64:
		parsed, err := money.FromFloat64(float64(val))
		if err == nil {
			return parsed
		}
	case []byte:
		return parseMoneyString(string(val))
	case string:
		return parseMoneyString(val)
	default:
		return parseMoneyString(fmt.Sprintf("%v", val))
	}
	return 0
}

func parseMoneyString(s string) money.Amount {
	if s == "" {
		return 0
	}
	value, err := money.FromString(s)
	if err != nil {
		logger.Global().Warn("failed to parse bank account balance sum from database",
			logger.String("value", s),
			logger.Error(err),
		)
		return 0
	}
	return value
}
