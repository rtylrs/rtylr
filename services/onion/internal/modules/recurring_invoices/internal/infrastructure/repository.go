package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/recurring_invoices/internal/domain"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{pool: dbclient.Pool()}
}

func (r *Repository) List(ctx context.Context, ownerID int64) ([]domain.RecurringInvoice, error) {
	query := `
		SELECT id, owner_id, template_invoice_id, frequency, next_invoice_date, end_date, is_active, created_at
		FROM finance_recurring_invoices
		WHERE owner_id = ?
		ORDER BY next_invoice_date ASC`

	rows, err := r.pool.QueryContext(ctx, query, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list recurring invoices", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	var items []domain.RecurringInvoice
	for rows.Next() {
		var item domain.RecurringInvoice
		if err := item.FromRow(rows); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *Repository) Create(ctx context.Context, ownerID int64, input domain.RecurringInvoiceInput) (int64, error) {
	query := `
		INSERT INTO finance_recurring_invoices
		(owner_id, template_invoice_id, frequency, next_invoice_date, end_date, is_active)
		VALUES (?, ?, ?, ?, ?, ?)`

	var endDate sql.NullString
	if input.EndDate != "" {
		endDate = sql.NullString{String: input.EndDate, Valid: true}
	}

	res, err := r.pool.ExecContext(ctx, query,
		ownerID, input.TemplateInvoiceID, input.Frequency,
		input.NextInvoiceDate, endDate, input.IsActive,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create recurring invoice", logger.Error(err))
		return 0, err
	}

	return res.LastInsertId()
}

func (r *Repository) Update(ctx context.Context, ownerID, id int64, input domain.RecurringInvoiceInput) error {
	query := `
		UPDATE finance_recurring_invoices
		SET template_invoice_id = ?, frequency = ?, next_invoice_date = ?, end_date = ?, is_active = ?
		WHERE id = ? AND owner_id = ?`

	var endDate sql.NullString
	if input.EndDate != "" {
		endDate = sql.NullString{String: input.EndDate, Valid: true}
	}

	_, err := r.pool.ExecContext(ctx, query,
		input.TemplateInvoiceID, input.Frequency, input.NextInvoiceDate,
		endDate, input.IsActive, id, ownerID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update recurring invoice", logger.Error(err))
		return err
	}

	return nil
}
