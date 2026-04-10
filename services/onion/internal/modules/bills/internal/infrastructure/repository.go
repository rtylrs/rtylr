package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/bills/internal/domain"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{pool: dbclient.Pool()}
}

func (r *Repository) List(ctx context.Context, ownerID int64) ([]domain.Bill, error) {
	query := `
		SELECT id, owner_id, supplier_id, amount, currency, due_date, status, description, created_at, updated_at
		FROM finance_bills
		WHERE owner_id = ?
		ORDER BY due_date ASC`

	rows, err := r.pool.QueryContext(ctx, query, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list bills", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	var items []domain.Bill
	for rows.Next() {
		var item domain.Bill
		if err := item.FromRow(rows); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *Repository) Create(ctx context.Context, ownerID int64, input domain.BillInput) (int64, error) {
	query := `
		INSERT INTO finance_bills (owner_id, supplier_id, amount, currency, due_date, status, description)
		VALUES (?, ?, ?, ?, ?, ?, ?)`

	status := input.Status
	if status == "" {
		status = "pending"
	}
	currency := input.Currency
	if currency == "" {
		currency = "USD"
	}

	res, err := r.pool.ExecContext(ctx, query,
		ownerID,
		sql.NullInt64{Int64: input.SupplierID, Valid: input.SupplierID != 0},
		input.Amount, currency, input.DueDate, status,
		sql.NullString{String: input.Description, Valid: input.Description != ""},
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create bill", logger.Error(err))
		return 0, err
	}

	return res.LastInsertId()
}

func (r *Repository) Update(ctx context.Context, ownerID, id int64, input domain.BillInput) error {
	query := `
		UPDATE finance_bills
		SET supplier_id = ?, amount = ?, currency = ?, due_date = ?, status = ?, description = ?
		WHERE id = ? AND owner_id = ?`

	_, err := r.pool.ExecContext(ctx, query,
		sql.NullInt64{Int64: input.SupplierID, Valid: input.SupplierID != 0},
		input.Amount, input.Currency, input.DueDate, input.Status,
		sql.NullString{String: input.Description, Valid: input.Description != ""},
		id, ownerID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update bill", logger.Error(err))
		return err
	}

	return nil
}
