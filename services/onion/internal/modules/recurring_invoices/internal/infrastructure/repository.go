package infrastructure

import (
	"context"
	"database/sql"
	"time"

	"github.com/voxire/onion/internal/modules/recurring_invoices/internal/domain"
	"github.com/voxire/onion/internal/modules/recurring_invoices/internal/infrastructure/persistence/sqlc"
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

func (r *Repository) List(ctx context.Context, businessID int64) ([]domain.RecurringInvoice, error) {
	rows, err := r.queries.ListRecurringInvoices(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list recurring invoices", logger.Error(err))
		return nil, err
	}

	items := make([]domain.RecurringInvoice, 0, len(rows))
	for _, row := range rows {
		ri := domain.RecurringInvoice{
			ID:                row.ID,
			OwnerID:           row.OwnerID,
			BusinessID:        row.BusinessID,
			TemplateInvoiceID: row.TemplateInvoiceID,
			Frequency:         string(row.Frequency),
			NextInvoiceDate:   row.NextInvoiceDate,
			IsActive:          nullBool(row.IsActive),
			CreatedAt:         nullTime(row.CreatedAt),
		}
		if row.EndDate.Valid {
			ri.EndDate = row.EndDate.Time
		}
		items = append(items, ri)
	}

	return items, nil
}

func (r *Repository) Create(ctx context.Context, businessID int64, input domain.RecurringInvoiceInput) (int64, error) {
	nextDate, err := time.Parse("2006-01-02", input.NextInvoiceDate)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to parse next_invoice_date", logger.Error(err))
		return 0, err
	}

	res, err := r.queries.CreateRecurringInvoice(ctx, sqlc.CreateRecurringInvoiceParams{
		OwnerID:           businessID,
		BusinessID:        businessID,
		TemplateInvoiceID: input.TemplateInvoiceID,
		Frequency:         sqlc.FinanceRecurringInvoicesFrequency(input.Frequency),
		NextInvoiceDate:   nextDate,
		EndDate:           nullTimeStr(input.EndDate),
		IsActive:          sql.NullBool{Bool: input.IsActive, Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create recurring invoice", logger.Error(err))
		return 0, err
	}

	return res.LastInsertId()
}

func (r *Repository) Update(ctx context.Context, businessID, id int64, input domain.RecurringInvoiceInput) error {
	nextDate, err := time.Parse("2006-01-02", input.NextInvoiceDate)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to parse next_invoice_date", logger.Error(err))
		return err
	}

	err = r.queries.UpdateRecurringInvoice(ctx, sqlc.UpdateRecurringInvoiceParams{
		TemplateInvoiceID: input.TemplateInvoiceID,
		Frequency:         sqlc.FinanceRecurringInvoicesFrequency(input.Frequency),
		NextInvoiceDate:   nextDate,
		EndDate:           nullTimeStr(input.EndDate),
		IsActive:          sql.NullBool{Bool: input.IsActive, Valid: true},
		ID:                id,
		BusinessID:        businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update recurring invoice", logger.Error(err))
		return err
	}

	return nil
}

func (r *Repository) SetActive(ctx context.Context, businessID, id int64, active bool) error {
	res, err := r.queries.SetRecurringInvoiceActive(ctx, sqlc.SetRecurringInvoiceActiveParams{
		IsActive:   sql.NullBool{Bool: active, Valid: true},
		ID:         id,
		BusinessID: businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update recurring invoice active flag", logger.Error(err))
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrRecurringInvoiceNotFound
	}
	return nil
}

// ---------------------------------------------------------------------------
// Conversion helpers
// ---------------------------------------------------------------------------

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

func nullTimeStr(s string) sql.NullTime {
	if s == "" {
		return sql.NullTime{}
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: t, Valid: true}
}
