package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"github.com/voxire/onion/internal/modules/finance/internal/outbox"
	"time"

	"github.com/Masterminds/squirrel"

	crmcontract "github.com/voxire/onion/internal/contracts/crm"
	"github.com/voxire/onion/internal/modules/finance/invoices/internal/domain"
	"github.com/voxire/onion/internal/modules/finance/invoices/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/eventbus"
	"github.com/voxire/onion/internal/platform/logger"
)

// dynamic-query: squirrel
func createInvoiceItems(ctx context.Context, db squirrel.BaseRunner, invoiceID int64, items []domain.InvoiceItem) error {
	if len(items) == 0 {
		return nil
	}

	ib := squirrel.Insert("finance_invoice_items").
		Columns("invoice_id", "description", "quantity", "unit", "unit_price", "tax_rate", "discount_pct", "total", "sort_order")
	for _, item := range items {
		ib = ib.Values(
			invoiceID,
			item.Description,
			fmtFloat(item.Quantity),
			sqlNullString(item.Unit),
			fmtFloat(item.UnitPrice),
			sqlNullString(fmtFloat(item.TaxRate)),
			sqlNullString(fmtFloat(item.DiscountPct)),
			fmtFloat(item.Total),
			sql.NullInt32{Int32: int32(item.SortOrder), Valid: true}, // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		)
	}

	_, err := ib.PlaceholderFormat(squirrel.Question).RunWith(db).ExecContext(ctx)
	return err
}

func (r *Repository) GetInvoiceItems(ctx context.Context, invoiceID int64) ([]domain.InvoiceItem, error) {
	rows, err := r.queries.GetInvoiceItems(ctx, invoiceID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get invoice items", logger.Error(err))

		return nil, err
	}

	items := make([]domain.InvoiceItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, domain.InvoiceItem{
			ID:          row.ID,
			InvoiceID:   row.InvoiceID,
			Description: row.Description,
			Quantity:    parseFloat(row.Quantity),
			Unit:        nullStr(row.Unit, "unit"),
			UnitPrice:   parseFloat(row.UnitPrice),
			TaxRate:     parseFloatNull(row.TaxRate),
			DiscountPct: parseFloatNull(row.DiscountPct),
			Total:       parseFloat(row.Total),
			SortOrder:   int(toInt32(row.SortOrder)),
		})
	}

	return items, nil
}

// CreateInvoiceItems is retained for legacy callers; application writes use
// CreateInvoiceWithItems to avoid a split transaction.
func (r *Repository) CreateInvoiceItems(ctx context.Context, invoiceID int64, items []domain.InvoiceItem) error {
	return createInvoiceItems(ctx, r.pool, invoiceID, items)
}

func (r *Repository) DeleteInvoiceItems(ctx context.Context, invoiceID int64) error {
	err := r.queries.DeleteInvoiceItems(ctx, invoiceID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete invoice items", logger.Error(err))
	}

	return err
}

func (r *Repository) UpdateInvoiceWithItems(ctx context.Context, businessID int64, inv domain.Invoice, items []domain.InvoiceItem) (err error) {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to begin invoice update transaction", logger.Error(err))
		return err
	}
	// Rollback is a no-op after a successful commit and also fires on panic,
	// so an interrupted transaction never lingers on the pooled connection.
	defer func() { _ = tx.Rollback() }()

	qtx := r.queries.WithTx(tx)
	status, err := qtx.GetInvoiceStatusForUpdate(ctx, sqlc.GetInvoiceStatusForUpdateParams{
		ID:         inv.ID,
		BusinessID: businessID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrInvoiceNotFound
	}
	if err != nil {
		return err
	}
	if !status.Valid || string(status.FinanceInvoicesStatus) != domain.StatusDraft {
		return domain.ErrInvoiceNotDraft
	}

	err = qtx.UpdateInvoice(ctx, sqlc.UpdateInvoiceParams{
		ClientName:     inv.ClientName,
		ClientEmail:    sqlNullString(inv.ClientEmail),
		ClientAddress:  sqlNullString(inv.ClientAddress),
		ClientTaxID:    sqlNullString(inv.ClientTaxID),
		CustomerID:     nullableInt64FromPtr(inv.CustomerID),
		IssueDate:      inv.IssueDate,
		DueDate:        inv.DueDate,
		Currency:       sqlNullString(inv.Currency),
		Subtotal:       sqlNullString(fmtFloat(inv.Subtotal)),
		DiscountType:   sqlc.NullFinanceInvoicesDiscountType{FinanceInvoicesDiscountType: sqlc.FinanceInvoicesDiscountType(inv.DiscountType), Valid: inv.DiscountType != ""},
		DiscountValue:  sqlNullString(fmtFloat(inv.DiscountValue)),
		DiscountAmount: sqlNullString(fmtFloat(inv.DiscountAmount)),
		TaxRate:        sqlNullString(fmtFloat(inv.TaxRate)),
		TaxAmount:      sqlNullString(fmtFloat(inv.TaxAmount)),
		Total:          sqlNullString(fmtFloat(inv.Total)),
		AmountDue:      sqlNullString(fmtFloat(inv.AmountDue)),
		Notes:          sqlNullString(inv.Notes),
		Terms:          sqlNullString(inv.Terms),
		ID:             inv.ID,
		BusinessID:     businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update invoice", logger.Error(err))
		return err
	}

	if items != nil {
		if err = qtx.DeleteInvoiceItems(ctx, inv.ID); err != nil {
			return err
		}
		if err = createInvoiceItems(ctx, tx, inv.ID, items); err != nil {
			return err
		}
	}
	contact := crmcontract.CustomerContactCaptured{Name: inv.ClientName, Email: inv.ClientEmail}
	if err = outbox.Publish(ctx, tx, eventbus.Event{
		BusinessID:  businessID,
		Type:        crmcontract.CustomerContactCapturedType,
		SubjectType: "finance_invoice",
		SubjectID:   inv.ID,
		Payload:     contact,
		DedupeKey:   crmcontract.CustomerContactCaptureDedupeKey("finance_invoice", inv.ID, contact),
	}); err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		logger.Ctx(ctx).Error("Failed to commit invoice update transaction", logger.Error(err))
		return err
	}

	return nil
}

// UpdateInvoice retains the legacy header-only repository operation. The
// application uses UpdateInvoiceWithItems for atomic header/item updates.
func (r *Repository) UpdateInvoice(ctx context.Context, businessID int64, inv domain.Invoice) error {
	return r.UpdateInvoiceWithItems(ctx, businessID, inv, nil)
}

func (r *Repository) UpdateInvoiceStatus(ctx context.Context, businessID, invoiceID int64, status string) error {
	// The UPDATE below refuses the same values, but a status-only write must
	// not be able to fabricate a money-bearing state even if a caller bypasses
	// the application guard.
	if err := domain.ValidateSettableInvoiceStatus(status); err != nil {
		return err
	}
	_, err := r.queries.UpdateInvoiceStatus(ctx, sqlc.UpdateInvoiceStatusParams{
		Status:     sqlc.NullFinanceInvoicesStatus{FinanceInvoicesStatus: sqlc.FinanceInvoicesStatus(status), Valid: status != ""},
		ID:         invoiceID,
		BusinessID: businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update invoice status", logger.Error(err))
	}

	return err
}

func (r *Repository) UpdateInvoicePaymentTotals(ctx context.Context, businessID, invoiceID int64, amountPaid, amountDue float64, status string, paidDate *time.Time) error {
	if paidDate != nil {
		err := r.queries.UpdateInvoicePaymentTotalsWithPaidDate(ctx, sqlc.UpdateInvoicePaymentTotalsWithPaidDateParams{
			AmountPaid: sqlNullString(fmtFloat(amountPaid)),
			AmountDue:  sqlNullString(fmtFloat(amountDue)),
			Status:     sqlc.NullFinanceInvoicesStatus{FinanceInvoicesStatus: sqlc.FinanceInvoicesStatus(status), Valid: status != ""},
			PaidDate:   sql.NullTime{Time: *paidDate, Valid: true},
			ID:         invoiceID,
			BusinessID: businessID,
		})
		if err != nil {
			logger.Ctx(ctx).Error("Failed to update invoice payment totals", logger.Error(err))
		}

		return err
	}

	err := r.queries.UpdateInvoicePaymentTotals(ctx, sqlc.UpdateInvoicePaymentTotalsParams{
		AmountPaid: sqlNullString(fmtFloat(amountPaid)),
		AmountDue:  sqlNullString(fmtFloat(amountDue)),
		Status:     sqlc.NullFinanceInvoicesStatus{FinanceInvoicesStatus: sqlc.FinanceInvoicesStatus(status), Valid: status != ""},
		ID:         invoiceID,
		BusinessID: businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update invoice payment totals", logger.Error(err))
	}

	return err
}

func (r *Repository) DeleteInvoice(ctx context.Context, businessID, invoiceID int64) error {
	err := r.queries.DeleteInvoice(ctx, sqlc.DeleteInvoiceParams{
		ID:         invoiceID,
		BusinessID: businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete invoice", logger.Error(err))
	}

	return err
}
