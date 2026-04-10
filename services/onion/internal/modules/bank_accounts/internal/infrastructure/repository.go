package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/bank_accounts/internal/domain"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{pool: dbclient.Pool()}
}

func (r *Repository) List(ctx context.Context, ownerID int64) ([]domain.BankAccount, error) {
	query := `
		SELECT id, owner_id, name, institution, account_type, balance, currency, is_connected, last_synced_at, created_at
		FROM finance_bank_accounts
		WHERE owner_id = ?
		ORDER BY name ASC`

	rows, err := r.pool.QueryContext(ctx, query, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list bank accounts", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	var items []domain.BankAccount
	for rows.Next() {
		var item domain.BankAccount
		if err := item.FromRow(rows); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *Repository) Create(ctx context.Context, ownerID int64, input domain.BankAccountInput) (int64, error) {
	query := `
		INSERT INTO finance_bank_accounts (owner_id, name, institution, account_type, balance, currency, is_connected)
		VALUES (?, ?, ?, ?, ?, ?, ?)`

	currency := input.Currency
	if currency == "" {
		currency = "USD"
	}

	res, err := r.pool.ExecContext(ctx, query,
		ownerID, input.Name,
		sql.NullString{String: input.Institution, Valid: input.Institution != ""},
		sql.NullString{String: input.AccountType, Valid: input.AccountType != ""},
		input.Balance, currency, input.IsConnected,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create bank account", logger.Error(err))
		return 0, err
	}

	return res.LastInsertId()
}
