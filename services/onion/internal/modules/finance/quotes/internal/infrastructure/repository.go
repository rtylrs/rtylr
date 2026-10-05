package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"github.com/voxire/onion/internal/modules/finance/internal/outbox"

	crmcontract "github.com/voxire/onion/internal/contracts/crm"
	"github.com/voxire/onion/internal/modules/finance/quotes/internal/domain"
	"github.com/voxire/onion/internal/modules/finance/quotes/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/eventbus"
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

func (r *Repository) List(ctx context.Context, businessID int64, limit, offset int32) ([]domain.Quote, error) {
	rows, err := r.queries.ListQuotes(ctx, sqlc.ListQuotesParams{
		BusinessID: businessID,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list quotes", logger.Error(err))
		return nil, err
	}

	items := make([]domain.Quote, 0, len(rows))
	for _, row := range rows {
		items = append(items, quoteFromSqlc(row))
	}

	return items, nil
}

func (r *Repository) Create(ctx context.Context, businessID int64, input domain.QuoteInput) (int64, error) {
	status := input.Status
	if status == "" {
		status = "draft"
	}

	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to begin quote transaction", logger.Error(err))
		return 0, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	res, err := r.queries.WithTx(tx).CreateQuote(ctx, sqlc.CreateQuoteParams{
		OwnerID:       businessID,
		BusinessID:    businessID,
		CustomerName:  input.CustomerName,
		CustomerEmail: nullableString(input.CustomerEmail),
		Status:        sqlc.NullFinanceQuotesStatus{FinanceQuotesStatus: sqlc.FinanceQuotesStatus(status), Valid: true},
		Items:         rawJSON(input.Items),
		Subtotal:      nullableString(fmtFloat(input.Subtotal)),
		Tax:           nullableString(fmtFloat(input.Tax)),
		Discount:      nullableString(fmtFloat(input.Discount)),
		Total:         nullableString(fmtFloat(input.Total)),
		ValidUntil:    nullTimeStr(input.ValidUntil),
		Notes:         nullableString(input.Notes),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create quote", logger.Error(err))
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	contact := crmcontract.CustomerContactCaptured{Name: input.CustomerName, Email: input.CustomerEmail}
	if err = outbox.Publish(ctx, tx, eventbus.Event{
		BusinessID:  businessID,
		Type:        crmcontract.CustomerContactCapturedType,
		SubjectType: "finance_quote",
		SubjectID:   id,
		Payload:     contact,
		DedupeKey:   crmcontract.CustomerContactCaptureDedupeKey("finance_quote", id, contact),
	}); err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		logger.Ctx(ctx).Error("Failed to commit quote transaction", logger.Error(err))
		return 0, err
	}
	return id, nil
}

func (r *Repository) Update(ctx context.Context, businessID, id int64, input domain.QuoteInput) error {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to begin quote update transaction", logger.Error(err))
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	quoteRow, err := r.queries.WithTx(tx).GetQuote(ctx, sqlc.GetQuoteParams{ID: id, BusinessID: businessID})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrQuoteNotFound
		}
		return err
	}
	status := string(quoteRow.Status.FinanceQuotesStatus)
	if status != "draft" && status != "sent" {
		return domain.ErrQuoteNotEditable
	}

	result, err := r.queries.WithTx(tx).UpdateQuote(ctx, sqlc.UpdateQuoteParams{
		CustomerName:  input.CustomerName,
		CustomerEmail: nullableString(input.CustomerEmail),
		Status:        sqlc.NullFinanceQuotesStatus{FinanceQuotesStatus: sqlc.FinanceQuotesStatus(input.Status), Valid: input.Status != ""},
		Items:         rawJSON(input.Items),
		Subtotal:      nullableString(fmtFloat(input.Subtotal)),
		Tax:           nullableString(fmtFloat(input.Tax)),
		Discount:      nullableString(fmtFloat(input.Discount)),
		Total:         nullableString(fmtFloat(input.Total)),
		ValidUntil:    nullTimeStr(input.ValidUntil),
		Notes:         nullableString(input.Notes),
		ID:            id,
		BusinessID:    businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update quote", logger.Error(err))
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return domain.ErrQuoteNotEditable
	}
	contact := crmcontract.CustomerContactCaptured{Name: input.CustomerName, Email: input.CustomerEmail}
	if err = outbox.Publish(ctx, tx, eventbus.Event{
		BusinessID:  businessID,
		Type:        crmcontract.CustomerContactCapturedType,
		SubjectType: "finance_quote",
		SubjectID:   id,
		Payload:     contact,
		DedupeKey:   crmcontract.CustomerContactCaptureDedupeKey("finance_quote", id, contact),
	}); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		logger.Ctx(ctx).Error("Failed to commit quote update transaction", logger.Error(err))
		return err
	}
	committed = true

	return nil
}

func (r *Repository) UpdateStatus(ctx context.Context, businessID, id int64, status string) error {
	result, err := r.queries.UpdateQuoteStatus(ctx, sqlc.UpdateQuoteStatusParams{
		Status:     sqlc.NullFinanceQuotesStatus{FinanceQuotesStatus: sqlc.FinanceQuotesStatus(status), Valid: status != ""},
		ID:         id,
		BusinessID: businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update quote status", logger.Error(err))
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected != 0 {
		return nil
	}

	// MySQL reports changed rows, not matched rows, so re-setting the current
	// status also lands here. Re-read the row to tell that idempotent no-op
	// apart from a quote that does not exist in this business, mirroring
	// Delete below.
	if _, err := r.Get(ctx, businessID, id); err != nil {
		return err
	}

	return nil
}

func (r *Repository) Delete(ctx context.Context, businessID, id int64) error {
	result, err := r.queries.DeleteQuote(ctx, sqlc.DeleteQuoteParams{
		ID:         id,
		BusinessID: businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete quote", logger.Error(err))
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected != 0 {
		return nil
	}

	quote, err := r.Get(ctx, businessID, id)
	if errors.Is(err, domain.ErrQuoteNotFound) {
		return domain.ErrQuoteNotFound
	}
	if err != nil {
		return err
	}
	if !quote.IsEditable() {
		return domain.ErrQuoteNotEditable
	}

	return nil
}
