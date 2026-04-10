package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/quotes/internal/domain"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{pool: dbclient.Pool()}
}

func (r *Repository) List(ctx context.Context, ownerID int64) ([]domain.Quote, error) {
	query := `
		SELECT id, owner_id, customer_name, customer_email, status,
		       items, subtotal, tax, discount, total, valid_until, notes, created_at, updated_at
		FROM finance_quotes
		WHERE owner_id = ?
		ORDER BY created_at DESC`

	rows, err := r.pool.QueryContext(ctx, query, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list quotes", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	var items []domain.Quote
	for rows.Next() {
		var item domain.Quote
		if err := item.FromRow(rows); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *Repository) Create(ctx context.Context, ownerID int64, input domain.QuoteInput) (int64, error) {
	query := `
		INSERT INTO finance_quotes
		(owner_id, customer_name, customer_email, status, items, subtotal, tax, discount, total, valid_until, notes)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	status := input.Status
	if status == "" {
		status = "draft"
	}

	res, err := r.pool.ExecContext(ctx, query,
		ownerID, input.CustomerName,
		sql.NullString{String: input.CustomerEmail, Valid: input.CustomerEmail != ""},
		status,
		sql.NullString{String: input.Items, Valid: input.Items != ""},
		input.Subtotal, input.Tax, input.Discount, input.Total,
		sql.NullString{String: input.ValidUntil, Valid: input.ValidUntil != ""},
		sql.NullString{String: input.Notes, Valid: input.Notes != ""},
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create quote", logger.Error(err))
		return 0, err
	}

	return res.LastInsertId()
}

func (r *Repository) Update(ctx context.Context, ownerID, id int64, input domain.QuoteInput) error {
	query := `
		UPDATE finance_quotes
		SET customer_name = ?, customer_email = ?, status = ?, items = ?,
		    subtotal = ?, tax = ?, discount = ?, total = ?, valid_until = ?, notes = ?
		WHERE id = ? AND owner_id = ?`

	_, err := r.pool.ExecContext(ctx, query,
		input.CustomerName,
		sql.NullString{String: input.CustomerEmail, Valid: input.CustomerEmail != ""},
		input.Status,
		sql.NullString{String: input.Items, Valid: input.Items != ""},
		input.Subtotal, input.Tax, input.Discount, input.Total,
		sql.NullString{String: input.ValidUntil, Valid: input.ValidUntil != ""},
		sql.NullString{String: input.Notes, Valid: input.Notes != ""},
		id, ownerID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update quote", logger.Error(err))
		return err
	}

	return nil
}

func (r *Repository) Delete(ctx context.Context, ownerID, id int64) error {
	_, err := r.pool.ExecContext(ctx,
		`DELETE FROM finance_quotes WHERE id = ? AND owner_id = ?`, id, ownerID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete quote", logger.Error(err))
		return err
	}

	return nil
}
