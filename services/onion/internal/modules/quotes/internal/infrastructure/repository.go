package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/quotes/internal/domain"
	"github.com/voxire/onion/internal/modules/quotes/internal/infrastructure/persistence/sqlc"
	customersync "github.com/voxire/onion/internal/platform/customer_sync"
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

	if _, err = customersync.UpsertCRMCustomer(ctx, tx, businessID, customersync.Contact{
		Name:  input.CustomerName,
		Email: input.CustomerEmail,
	}); err != nil {
		logger.Ctx(ctx).Error("Failed to sync quote customer", logger.Error(err))
		return 0, err
	}

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

	if _, err = customersync.UpsertCRMCustomer(ctx, tx, businessID, customersync.Contact{
		Name:  input.CustomerName,
		Email: input.CustomerEmail,
	}); err != nil {
		logger.Ctx(ctx).Error("Failed to sync quote customer", logger.Error(err))
		return err
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

// ConvertToInvoice materialises a quote into a finance_invoices row and marks
// the quote as converted. The invoice insert is a cross-schema write: the
// quotes module owns finance_quotes and writes directly to finance_invoices
// here because the operation must be atomic. No cross-module Go imports are
// introduced — we use dbclient + raw SQL at the infrastructure layer only.
func (r *Repository) ConvertToInvoice(ctx context.Context, businessID, quoteID int64) (int64, error) {
	db := r.pool

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	qtx := r.queries.WithTx(tx)

	// Serialize invoice-number allocation per business. The business row is
	// the per-tenant mutex: locking it before the first read also starts this
	// transaction's read view after any competing conversion committed, so
	// CountInvoiceSequenceForQuote below cannot mint a number another
	// conversion has just taken.
	if _, err = qtx.LockBusinessForQuoteConversion(ctx, businessID); err != nil {
		_ = tx.Rollback()
		logger.Ctx(ctx).Error("Failed to lock business for quote conversion", logger.Error(err))
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

	issueDate := time.Now().UTC()
	// Fallback: quotes don't track an explicit due date, so default to
	// issue_date + 30 days. If the quote had a valid_until we use that.
	dueDate := issueDate.AddDate(0, 0, 30)
	if quoteRow.ValidUntil.Valid {
		dueDate = quoteRow.ValidUntil.Time
	}
	if dueDate.Before(issueDate) {
		dueDate = issueDate
	}
	sequence, err := qtx.CountInvoiceSequenceForQuote(ctx, sqlc.CountInvoiceSequenceForQuoteParams{
		BusinessID: businessID,
		CreatedAt:  sql.NullTime{Time: issueDate, Valid: true},
	})
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}
	invoiceNumber := fmt.Sprintf("INV-%d-%04d", issueDate.Year(), sequence+1)

	customerEmail := ""
	if quoteRow.CustomerEmail.Valid {
		customerEmail = quoteRow.CustomerEmail.String
	}
	notes := ""
	if quoteRow.Notes.Valid {
		notes = quoteRow.Notes.String
	}

	if _, err = customersync.UpsertCRMCustomer(ctx, tx, businessID, customersync.Contact{
		Name:  quoteRow.CustomerName,
		Email: customerEmail,
	}); err != nil {
		_ = tx.Rollback()
		logger.Ctx(ctx).Error("Failed to sync quote customer before conversion", logger.Error(err))
		return 0, err
	}

	// cross-schema write from quotes to finance_invoices — materialise quote
	// header row. Column list matches finance_invoices schema.
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

	// cross-schema write from quotes to finance_invoice_items — copy each
	// quote line as an invoice item. Quote items are stored as JSON so we
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
			total := floatFromAny(item["total"])
			if total == 0 {
				total = qty * unitPrice
			}
			// cross-schema write from quotes to finance_invoice_items
			if itemErr := qtx.InsertInvoiceItemFromQuote(ctx, sqlc.InsertInvoiceItemFromQuoteParams{
				InvoiceID:   invoiceID,
				Description: desc,
				Quantity:    fmtFloat(qty),
				Unit:        nullableStringInput("unit"),
				UnitPrice:   fmtFloat(unitPrice),
				TaxRate:     nullableStringInput(fmtFloat(0)),
				DiscountPct: nullableStringInput(fmtFloat(0)),
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

// CountByStatus returns the number of quotes in each status for the business.
func (r *Repository) CountByStatus(ctx context.Context, businessID int64) (map[string]int64, error) {
	rows, err := r.queries.CountQuotesByStatus(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to count quotes by status", logger.Error(err))
		return nil, err
	}

	counts := make(map[string]int64)
	for _, row := range rows {
		counts[string(row.Status.FinanceQuotesStatus)] = row.Cnt
	}

	return counts, nil
}

func nullableStringInput(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

func nullStrValue(s sql.NullString) sql.NullString { return s }

func stringFromAny(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func floatFromAny(v any) float64 {
	if v == nil {
		return 0
	}
	switch t := v.(type) {
	case float64:
		return t
	case float32:
		return float64(t)
	case int:
		return float64(t)
	case int64:
		return float64(t)
	case string:
		f, err := strconv.ParseFloat(t, 64)
		if err != nil {
			return 0
		}
		return f
	}
	return 0
}

// ---------------------------------------------------------------------------
// Mapping helpers
// ---------------------------------------------------------------------------

func quoteFromSqlc(row sqlc.FinanceQuote) domain.Quote {
	q := domain.Quote{
		ID:            row.ID,
		OwnerID:       row.OwnerID,
		BusinessID:    row.BusinessID,
		CustomerName:  row.CustomerName,
		CustomerEmail: nullStr(row.CustomerEmail),
		Status:        string(row.Status.FinanceQuotesStatus),
		Items:         string(row.Items),
		Subtotal:      parseFloatNull(row.Subtotal),
		Tax:           parseFloatNull(row.Tax),
		Discount:      parseFloatNull(row.Discount),
		Total:         parseFloatNull(row.Total),
		Notes:         nullStr(row.Notes),
		CreatedAt:     nullTime(row.CreatedAt),
		UpdatedAt:     nullTime(row.UpdatedAt),
	}
	if row.ValidUntil.Valid {
		q.ValidUntil = row.ValidUntil.Time
	}
	return q
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

func fmtFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', 2, 64)
}

func parseFloatNull(s sql.NullString) float64 {
	if !s.Valid || s.String == "" {
		return 0
	}
	v, err := strconv.ParseFloat(s.String, 64)
	if err != nil {
		logger.Global().Warn("failed to parse decimal from database", logger.String("value", s.String), logger.Error(err))
		return 0
	}
	return v
}

func rawJSON(s string) json.RawMessage {
	if s == "" {
		return json.RawMessage("[]")
	}
	return json.RawMessage(s)
}
