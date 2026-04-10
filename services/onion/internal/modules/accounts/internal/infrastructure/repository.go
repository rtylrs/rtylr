package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/accounts/internal/domain"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{pool: dbclient.Pool()}
}

func (r *Repository) List(ctx context.Context, ownerID int64) ([]domain.Account, error) {
	query := `
		SELECT id, owner_id, name, type, currency, balance, is_active, created_at
		FROM finance_accounts
		WHERE owner_id = ?
		ORDER BY type ASC, name ASC`

	rows, err := r.pool.QueryContext(ctx, query, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list accounts", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	var items []domain.Account
	for rows.Next() {
		var item domain.Account
		if err := item.FromRow(rows); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *Repository) Create(ctx context.Context, ownerID int64, input domain.AccountInput) (int64, error) {
	query := `
		INSERT INTO finance_accounts (owner_id, name, type, currency, balance, is_active)
		VALUES (?, ?, ?, ?, ?, ?)`

	currency := input.Currency
	if currency == "" {
		currency = "USD"
	}

	res, err := r.pool.ExecContext(ctx, query,
		ownerID, input.Name, input.Type, currency, input.Balance, input.IsActive,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create account", logger.Error(err))
		return 0, err
	}

	return res.LastInsertId()
}
