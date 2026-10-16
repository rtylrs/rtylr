package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/voxire/onion/internal/modules/finance/internal/outbox"
	"math"
	"time"

	crmcontract "github.com/voxire/onion/internal/contracts/crm"
	"github.com/voxire/onion/internal/modules/finance/internal/invoicenumber"
	"github.com/voxire/onion/internal/modules/finance/quotes/internal/domain"
	"github.com/voxire/onion/internal/modules/finance/quotes/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/eventbus"
	"github.com/voxire/onion/internal/platform/logger"
)

func (r *Repository) Get(ctx context.Context, businessID, id int64) (domain.Quote, error) {
	row, err := r.queries.GetQuote(ctx, sqlc.GetQuoteParams{ID: id, BusinessID: businessID})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Quote{}, domain.ErrQuoteNotFound
		}
		logger.Ctx(ctx).Error("Failed to get quote", logger.Error(err))
		return domain.Quote{}, err
	}
	return quoteFromSqlc(row), nil
}

// ConvertToInvoice commits the invoice and converted quote in one Finance transaction.
func (r *Repository) ConvertToInvoice(ctx context.Context, businessID, quoteID int64) (int64, error) {
	db := r.pool

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	qtx := r.queries.WithTx(tx)

	issueDate := time.Now().UTC()
	invoiceNumber, err := invoicenumber.Allocate(ctx, tx, businessID, issueDate.Year())
	if err != nil {
		return 0, err
	}

	quoteRow, err := qtx.GetQuote(ctx, sqlc.GetQuoteParams{ID: quoteID, BusinessID: businessID})
	if err != nil {
		_ = tx.Rollback()
		if errors.Is(err, sql.ErrNoRows) {
			return 0, domain.ErrQuoteNotFound
		}
		logger.Ctx(ctx).Error("Failed to load quote for conversion", logger.Error(err))
		return 0, err
	}

	if string(quoteRow.Status.FinanceQuotesStatus) == "converted" {
		_ = tx.Rollback()
		return 0, domain.ErrQuoteAlreadyConverted
	}

	// Fallback: quotes don't track an explicit due date, so default to
	// issue_date + 30 days. If the quote had a valid_until we use that.
	dueDate := issueDate.AddDate(0, 0, 30)
	if quoteRow.ValidUntil.Valid {
		dueDate = quoteRow.ValidUntil.Time
	}
	if dueDate.Before(issueDate) {
		dueDate = issueDate
	}

	customerEmail := ""
	if quoteRow.CustomerEmail.Valid {
		customerEmail = quoteRow.CustomerEmail.String
	}
	notes := ""
	if quoteRow.Notes.Valid {
		notes = quoteRow.Notes.String
	}

	// Materialize the invoice header from the stored quote.
	// Totals are client-supplied on quote write paths, so re-verify the
	// stored header is internally consistent before materialising it into
	// finance_invoices. A tampered or legacy-bad row must not become an
	// accounts-receivable document.
	subtotal := parseFloatNull(quoteRow.Subtotal)
	tax := parseFloatNull(quoteRow.Tax)
	discount := parseFloatNull(quoteRow.Discount)
	total := parseFloatNull(quoteRow.Total)
	if math.Abs(total-(subtotal-discount+tax)) > 0.005 {
		_ = tx.Rollback()
		logger.Ctx(ctx).Error("Refusing to convert quote with inconsistent totals",
			logger.Any("quoteId", quoteID),
			logger.Any("total", total),
		)
		return 0, domain.ErrQuoteTotalMismatch
	}

	res, err := qtx.InsertInvoiceFromQuote(ctx, sqlc.InsertInvoiceFromQuoteParams{
		OwnerID:        businessID,
		BusinessID:     businessID,
		InvoiceNumber:  invoiceNumber,
		ClientName:     quoteRow.CustomerName,
		ClientEmail:    nullableStringInput(customerEmail),
		Status:         sqlc.NullFinanceInvoicesStatus{FinanceInvoicesStatus: sqlc.FinanceInvoicesStatusDraft, Valid: true},
		IssueDate:      issueDate,
		DueDate:        dueDate,
		Subtotal:       nullStrValue(quoteRow.Subtotal),
		TaxAmount:      nullStrValue(quoteRow.Tax),
		DiscountAmount: nullStrValue(quoteRow.Discount),
		Total:          nullStrValue(quoteRow.Total),
		AmountDue:      nullStrValue(quoteRow.Total),
		Notes:          nullableStringInput(notes),
	})
	if err != nil {
		_ = tx.Rollback()
		logger.Ctx(ctx).Error("Failed to create invoice from quote", logger.Error(err))
		return 0, err
	}

	invoiceID, err := res.LastInsertId()
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}
	contact := crmcontract.CustomerContactCaptured{Name: quoteRow.CustomerName, Email: customerEmail}
	if err = outbox.Publish(ctx, tx, eventbus.Event{
		BusinessID:  businessID,
		Type:        crmcontract.CustomerContactCapturedType,
		SubjectType: "finance_invoice_from_quote",
		SubjectID:   invoiceID,
		Payload:     contact,
		DedupeKey:   crmcontract.CustomerContactCaptureDedupeKey("finance_invoice_from_quote", invoiceID, contact),
	}); err != nil {
		_ = tx.Rollback()
		return 0, err
	}

	// Copy each quote line into an invoice item. Quote items are stored as JSON so we
	// parse and fan out. Unknown fields fall back to safe defaults.
	if len(quoteRow.Items) > 0 {
		var items []map[string]any
		if decodeErr := json.Unmarshal(quoteRow.Items, &items); decodeErr != nil {
			_ = tx.Rollback()
			logger.Ctx(ctx).Error("Failed to unmarshal quote items",
				logger.Error(decodeErr))
			return 0, fmt.Errorf("invalid JSON in quote items: %w", decodeErr)
		}
		for i, item := range items {
			desc := stringFromAny(item["description"])
			if desc == "" {
				desc = stringFromAny(item["name"])
			}
			qty := floatFromAny(item["quantity"])
			if qty == 0 {
				qty = 1
			}
			unitPrice := floatFromAny(item["unit_price"])
			if unitPrice == 0 {
				unitPrice = floatFromAny(item["price"])
			}
			taxRate := floatFromAny(item["tax_rate"])
			if taxRate == 0 {
				taxRate = floatFromAny(item["tax"])
			}
			discountPct := floatFromAny(item["discount_pct"])
			if discountPct == 0 {
				discountPct = floatFromAny(item["discount"])
			}
			total := floatFromAny(item["total"])
			if total == 0 {
				total = qty * unitPrice
			}
			if itemErr := qtx.InsertInvoiceItemFromQuote(ctx, sqlc.InsertInvoiceItemFromQuoteParams{
				InvoiceID:   invoiceID,
				Description: desc,
				Quantity:    fmtFloat(qty),
				Unit:        nullableStringInput("unit"),
				UnitPrice:   fmtFloat(unitPrice),
				TaxRate:     nullableStringInput(fmtFloat(taxRate)),
				DiscountPct: nullableStringInput(fmtFloat(discountPct)),
				Total:       fmtFloat(total),
				SortOrder:   sql.NullInt32{Int32: int32(i), Valid: true},
			}); itemErr != nil {
				_ = tx.Rollback()
				logger.Ctx(ctx).Error("Failed to insert invoice item from quote", logger.Error(itemErr))
				return 0, itemErr
			}
		}
	}

	// The predicate makes the status flip conditional: under concurrency a
	// second conversion's update matches zero rows once the first commits,
	// so exactly one invoice can ever be created from a quote.
	statusRes, statusErr := qtx.MarkQuoteConverted(ctx, sqlc.MarkQuoteConvertedParams{
		ID:         quoteID,
		BusinessID: businessID,
	})
	if statusErr != nil {
		_ = tx.Rollback()
		logger.Ctx(ctx).Error("Failed to mark quote converted", logger.Error(statusErr))
		return 0, statusErr
	}
	if rows, _ := statusRes.RowsAffected(); rows == 0 {
		_ = tx.Rollback()
		return 0, domain.ErrQuoteAlreadyConverted
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return invoiceID, nil
}
