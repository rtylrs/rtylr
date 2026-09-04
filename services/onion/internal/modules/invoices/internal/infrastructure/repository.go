package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/Masterminds/squirrel"

	"github.com/voxire/onion/internal/modules/invoices/internal/domain"
	"github.com/voxire/onion/internal/modules/invoices/internal/infrastructure/persistence/sqlc"
	customersync "github.com/voxire/onion/internal/platform/customer_sync"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/money"
	"github.com/voxire/onion/internal/platform/searchutil"
)

type Repository struct {
	queries *sqlc.Queries
	pool    *sql.DB
}

// CreateInvoice retains the legacy header-only repository operation used by
// integration fixtures. Application writes use CreateInvoiceWithItems so the
// header and items share one transaction.
func (r *Repository) CreateInvoice(ctx context.Context, businessID int64, inv domain.Invoice) (int64, error) {
	return r.CreateInvoiceWithItems(ctx, businessID, inv, nil)
}

func NewRepository(db *sql.DB) *Repository {
	pool := db

	return &Repository{
		queries: sqlc.New(pool),
		pool:    pool,
	}
}

// ---------------------------------------------------------------------------
// Invoices
// ---------------------------------------------------------------------------

// dynamic-query: squirrel
func (r *Repository) ListInvoices(ctx context.Context, businessID int64, status, search, dateFrom, dateTo string, limit, offset int) ([]domain.Invoice, error) {
	qb := squirrel.Select(
		"id", "owner_id", "business_id", "invoice_number", "client_name", "client_email", "client_address", "client_tax_id",
		"status", "issue_date", "due_date", "paid_date", "currency", "subtotal", "discount_type",
		"discount_value", "discount_amount", "tax_rate", "tax_amount", "total", "amount_paid", "amount_due",
		"notes", "terms", "created_at", "updated_at",
	).
		From("finance_invoices").
		Where(squirrel.Eq{"business_id": businessID})

	if status != "" {
		qb = qb.Where(squirrel.Eq{"status": status})
	}

	if search = strings.TrimSpace(search); search != "" {
		booleanQuery := searchutil.BooleanPrefixQuery(search)
		if booleanQuery != "" {
			qb = qb.Where(
				"(invoice_number = ? OR client_email = ? OR MATCH(invoice_number, client_name, client_email) AGAINST (? IN BOOLEAN MODE))",
				search,
				search,
				booleanQuery,
			)
		} else {
			qb = qb.Where(squirrel.Or{squirrel.Eq{"invoice_number": search}, squirrel.Eq{"client_email": search}})
		}
	}

	if dateFrom != "" {
		qb = qb.Where(squirrel.GtOrEq{"issue_date": dateFrom})
	}

	if dateTo != "" {
		qb = qb.Where(squirrel.LtOrEq{"issue_date": dateTo})
	}

	qb = qb.OrderBy("created_at DESC, id DESC").Limit(uint64(limit)).Offset(uint64(offset)) // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.

	rows, err := qb.PlaceholderFormat(squirrel.Question).RunWith(r.pool).QueryContext(ctx)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list invoices", logger.Error(err))

		return nil, err
	}
	defer rows.Close()

	invoices := make([]domain.Invoice, 0, limit)

	for rows.Next() {
		var inv domain.Invoice
		if scanErr := inv.FromRow(rows); scanErr != nil {
			logger.Ctx(ctx).Error("Failed to scan invoice row", logger.Error(scanErr))

			return nil, scanErr
		}

		invoices = append(invoices, inv)
	}

	return invoices, rows.Err()
}

func (r *Repository) ListUnbilledBillableTimeEntries(ctx context.Context, businessID int64, limit, offset int) ([]domain.BillableTimeEntry, error) {
	rows, err := r.queries.ListUnbilledBillableTimeEntries(ctx, sqlc.ListUnbilledBillableTimeEntriesParams{
		BusinessID: businessID,
		Limit:      int32(limit),  // #nosec G115 -- pagination is normalized by the application policy.
		Offset:     int32(offset), // #nosec G115 -- pagination is normalized by the application policy.
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to list unbilled billable time entries", logger.Any("business_id", businessID), logger.Error(err))
		return nil, err
	}

	entries := make([]domain.BillableTimeEntry, 0, len(rows))
	for _, row := range rows {
		entry := domain.BillableTimeEntry{
			ID:              row.ID,
			EmployeeID:      row.EmployeeID,
			Date:            row.Date.Format("2006-01-02"),
			DurationMinutes: int64(row.DurationMinutes),
			IsBillable:      row.IsBillable,
			Status:          row.Status,
			Source:          row.Source,
		}
		if row.ProjectID.Valid {
			projectID := row.ProjectID.Int64
			entry.ProjectID = &projectID
		}
		if row.TaskID.Valid {
			taskID := row.TaskID.Int64
			entry.TaskID = &taskID
		}
		if row.Description.Valid {
			entry.Description = row.Description.String
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func (r *Repository) CountUnbilledBillableTimeEntries(ctx context.Context, businessID int64) (int64, int64, error) {
	totals, err := r.queries.CountUnbilledBillableTimeEntries(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("failed to count unbilled billable time entries", logger.Any("business_id", businessID), logger.Error(err))
	}
	return totals.Total, totals.TotalMinutes, err
}

func (r *Repository) GetInvoiceByID(ctx context.Context, businessID, invoiceID int64) (*domain.Invoice, error) {
	row, err := r.queries.GetInvoiceByID(ctx, sqlc.GetInvoiceByIDParams{
		ID:         invoiceID,
		BusinessID: businessID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrInvoiceNotFound
		}
		logger.Ctx(ctx).Error("Failed to get invoice by ID", logger.Error(err))

		return nil, err
	}

	inv := invoiceFromSqlc(row)

	return &inv, nil
}

func (r *Repository) CreateInvoiceWithItems(ctx context.Context, businessID int64, inv domain.Invoice, items []domain.InvoiceItem) (int64, error) {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to begin invoice transaction", logger.Error(err))
		return 0, err
	}
	// Rollback is a no-op after a successful commit and also fires on panic,
	// so an interrupted transaction never lingers on the pooled connection.
	defer func() { _ = tx.Rollback() }()

	if _, err = customersync.UpsertCRMCustomer(ctx, tx, businessID, customersync.Contact{
		Name:  inv.ClientName,
		Email: inv.ClientEmail,
	}); err != nil {
		logger.Ctx(ctx).Error("Failed to sync invoice customer", logger.Error(err))
		return 0, err
	}

	result, err := r.queries.WithTx(tx).CreateInvoice(ctx, sqlc.CreateInvoiceParams{
		OwnerID:        businessID,
		BusinessID:     businessID,
		InvoiceNumber:  inv.InvoiceNumber,
		ClientName:     inv.ClientName,
		ClientEmail:    sqlNullString(inv.ClientEmail),
		ClientAddress:  sqlNullString(inv.ClientAddress),
		ClientTaxID:    sqlNullString(inv.ClientTaxID),
		CustomerID:     nullableInt64FromPtr(inv.CustomerID),
		Status:         sqlc.NullFinanceInvoicesStatus{FinanceInvoicesStatus: sqlc.FinanceInvoicesStatus(inv.Status), Valid: inv.Status != ""},
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
		AmountPaid:     sqlNullString(fmtFloat(inv.AmountPaid)),
		AmountDue:      sqlNullString(fmtFloat(inv.AmountDue)),
		Notes:          sqlNullString(inv.Notes),
		Terms:          sqlNullString(inv.Terms),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create invoice", logger.Error(err))

		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to read created invoice ID", logger.Error(err))

		return 0, err
	}
	if err = createInvoiceItems(ctx, tx, id, items); err != nil {
		logger.Ctx(ctx).Error("Failed to create invoice items", logger.Error(err))
		return 0, err
	}

	if err = tx.Commit(); err != nil {
		logger.Ctx(ctx).Error("Failed to commit invoice transaction", logger.Error(err))
		return 0, err
	}

	return id, nil
}

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

	if _, err = customersync.UpsertCRMCustomer(ctx, tx, businessID, customersync.Contact{
		Name:  inv.ClientName,
		Email: inv.ClientEmail,
	}); err != nil {
		logger.Ctx(ctx).Error("Failed to sync invoice customer", logger.Error(err))
		return err
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

func (r *Repository) CountInvoiceSequence(ctx context.Context, businessID int64, year int) (int64, error) {
	start, end := invoiceSequenceYearBounds(year)
	count, err := r.queries.CountInvoiceSequence(ctx, sqlc.CountInvoiceSequenceParams{
		BusinessID:     businessID,
		CreatedAtStart: sql.NullTime{Time: start, Valid: true},
		CreatedAtEnd:   sql.NullTime{Time: end, Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to count invoice sequence", logger.Error(err))

		return 0, err
	}

	return count, nil
}

func invoiceSequenceYearBounds(year int) (time.Time, time.Time) {
	start := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)
	return start, start.AddDate(1, 0, 0)
}

// ---------------------------------------------------------------------------
// Payments
// ---------------------------------------------------------------------------

func (r *Repository) CreatePayment(ctx context.Context, businessID int64, p domain.Payment) (int64, error) {
	result, err := r.queries.CreatePayment(ctx, sqlc.CreatePaymentParams{
		OwnerID:     businessID,
		BusinessID:  businessID,
		InvoiceID:   p.InvoiceID,
		Amount:      fmtFloat(p.Amount),
		Currency:    sqlNullString(p.Currency),
		PaymentDate: p.PaymentDate,
		Method:      sqlc.FinancePaymentsMethod(p.Method),
		Reference:   sqlNullString(p.Reference),
		Notes:       sqlNullString(p.Notes),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create payment", logger.Error(err))

		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to read created payment ID", logger.Error(err))

		return 0, err
	}

	return id, nil
}

// RecordPayment atomically locks the invoice, returns an existing referenced
// payment when present, or inserts the payment and updates invoice totals on
// the same transaction.
func (r *Repository) RecordPayment(ctx context.Context, businessID int64, input domain.PaymentInput) (*domain.Payment, error) {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	qtx := r.queries.WithTx(tx)
	row, err := qtx.GetInvoiceByIDForUpdate(ctx, sqlc.GetInvoiceByIDForUpdateParams{
		ID: input.InvoiceID, BusinessID: businessID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrInvoiceNotFound
	}
	if err != nil {
		return nil, err
	}
	inv := invoiceFromSqlc(row)

	if strings.TrimSpace(input.Reference) != "" {
		existing, lookupErr := qtx.GetPaymentByReference(ctx, sqlc.GetPaymentByReferenceParams{
			BusinessID: businessID,
			InvoiceID:  input.InvoiceID,
			Reference:  sqlNullString(input.Reference),
		})
		if lookupErr == nil {
			payment := paymentFromSqlc(existing)
			return &payment, nil
		}
		if !errors.Is(lookupErr, sql.ErrNoRows) {
			return nil, lookupErr
		}
	}
	if domain.IsTerminal(inv.Status) {
		return nil, domain.ErrInvoiceNotDraft
	}

	currency := input.Currency
	if currency == "" {
		currency = inv.Currency
	}
	p := domain.Payment{
		OwnerID: businessID, BusinessID: businessID, InvoiceID: input.InvoiceID,
		Amount: input.Amount, Currency: currency, PaymentDate: mustParsePaymentDate(input.PaymentDate),
		Method: input.Method, Reference: input.Reference, Notes: input.Notes,
	}
	result, err := qtx.CreatePayment(ctx, sqlc.CreatePaymentParams{
		OwnerID: businessID, BusinessID: businessID, InvoiceID: p.InvoiceID, Amount: fmtFloat(p.Amount),
		Currency: sqlNullString(p.Currency), PaymentDate: p.PaymentDate, Method: sqlc.FinancePaymentsMethod(p.Method),
		Reference: sqlNullString(p.Reference), Notes: sqlNullString(p.Notes),
	})
	if err != nil {
		return nil, err
	}
	paymentID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	p.ID = paymentID

	// Accumulate and compare in integer cents: float64 residue from parsed
	// DECIMAL strings could otherwise leave newAmountDue at +2.4e-15 and keep
	// a fully paid invoice in "partial" forever (or flip one early).
	paymentCents := paymentAmountToCents(p.Amount)
	totalCents := nullDecimalToCents(row.Total)
	paidCents := nullDecimalToCents(row.AmountPaid)
	newPaidCents := paidCents + paymentCents
	newDueCents := totalCents - newPaidCents
	if newDueCents < 0 {
		newDueCents = 0
	}
	newStatus := "partial"
	var paidDate *time.Time
	if newDueCents <= 0 {
		newStatus = "paid"
		paidDateValue := p.PaymentDate
		paidDate = &paidDateValue
	}
	if paidDate != nil {
		err = qtx.UpdateInvoicePaymentTotalsWithPaidDate(ctx, sqlc.UpdateInvoicePaymentTotalsWithPaidDateParams{
			AmountPaid: sqlNullString(newPaidCents.String()), AmountDue: sqlNullString(newDueCents.String()),
			Status:   sqlc.NullFinanceInvoicesStatus{FinanceInvoicesStatus: sqlc.FinanceInvoicesStatus(newStatus), Valid: true},
			PaidDate: sql.NullTime{Time: *paidDate, Valid: true}, ID: input.InvoiceID, BusinessID: businessID,
		})
	} else {
		err = qtx.UpdateInvoicePaymentTotals(ctx, sqlc.UpdateInvoicePaymentTotalsParams{
			AmountPaid: sqlNullString(newPaidCents.String()), AmountDue: sqlNullString(newDueCents.String()),
			Status: sqlc.NullFinanceInvoicesStatus{FinanceInvoicesStatus: sqlc.FinanceInvoicesStatus(newStatus), Valid: true}, ID: input.InvoiceID, BusinessID: businessID,
		})
	}
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &p, nil
}

func mustParsePaymentDate(value string) time.Time {
	parsed, _ := time.Parse("2006-01-02", value)
	return parsed
}

func paymentFromSqlc(row sqlc.FinancePayment) domain.Payment {
	return domain.Payment{
		ID: row.ID, OwnerID: row.OwnerID, BusinessID: row.BusinessID, InvoiceID: row.InvoiceID,
		Amount: parseFloat(row.Amount), Currency: nullStr(row.Currency, "USD"), PaymentDate: row.PaymentDate,
		Method: string(row.Method), Reference: nullStr(row.Reference, ""), Notes: nullStr(row.Notes, ""),
		CreatedAt: nullTime(row.CreatedAt),
	}
}

func (r *Repository) GetPaymentsByInvoice(ctx context.Context, invoiceID int64) ([]domain.Payment, error) {
	rows, err := r.queries.GetPaymentsByInvoice(ctx, invoiceID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get payments by invoice", logger.Error(err))

		return nil, err
	}

	payments := make([]domain.Payment, 0, len(rows))
	for _, row := range rows {
		payments = append(payments, domain.Payment{
			ID:          row.ID,
			OwnerID:     row.OwnerID,
			BusinessID:  row.BusinessID,
			InvoiceID:   row.InvoiceID,
			Amount:      parseFloat(row.Amount),
			Currency:    nullStr(row.Currency, "USD"),
			PaymentDate: row.PaymentDate,
			Method:      string(row.Method),
			Reference:   nullStr(row.Reference, ""),
			Notes:       nullStr(row.Notes, ""),
			CreatedAt:   nullTime(row.CreatedAt),
		})
	}

	return payments, nil
}

// ---------------------------------------------------------------------------
// Expenses
// ---------------------------------------------------------------------------

// dynamic-query: squirrel
func (r *Repository) ListExpenses(ctx context.Context, businessID, categoryID int64, search, dateFrom, dateTo string, limit, offset int) ([]domain.Expense, error) {
	qb := squirrel.Select(
		"id", "owner_id", "business_id", "category_id", "description", "amount", "currency", "expense_date", "vendor",
		"payment_method", "receipt_url", "is_tax_deductible", "tax_amount", "notes", "status", "created_at", "updated_at",
	).
		From("finance_expenses").
		Where(squirrel.Eq{"business_id": businessID})

	if categoryID > 0 {
		qb = qb.Where(squirrel.Eq{"category_id": categoryID})
	}

	if search = strings.TrimSpace(search); search != "" {
		booleanQuery := searchutil.BooleanPrefixQuery(search)
		if booleanQuery != "" {
			qb = qb.Where(
				"(vendor = ? OR MATCH(description, vendor) AGAINST (? IN BOOLEAN MODE))",
				search,
				booleanQuery,
			)
		} else {
			qb = qb.Where(squirrel.Eq{"vendor": search})
		}
	}

	if dateFrom != "" {
		qb = qb.Where(squirrel.GtOrEq{"expense_date": dateFrom})
	}

	if dateTo != "" {
		qb = qb.Where(squirrel.LtOrEq{"expense_date": dateTo})
	}

	qb = qb.OrderBy("expense_date DESC").Limit(uint64(limit)).Offset(uint64(offset)) // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.

	rows, err := qb.PlaceholderFormat(squirrel.Question).RunWith(r.pool).QueryContext(ctx)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list expenses", logger.Error(err))

		return nil, err
	}
	defer rows.Close()

	expenses := make([]domain.Expense, 0, limit)

	for rows.Next() {
		var e domain.Expense
		if scanErr := e.FromRow(rows); scanErr != nil {
			logger.Ctx(ctx).Error("Failed to scan expense row", logger.Error(scanErr))

			return nil, scanErr
		}

		expenses = append(expenses, e)
	}

	return expenses, rows.Err()
}

func (r *Repository) GetExpenseByID(ctx context.Context, businessID, expenseID int64) (*domain.Expense, error) {
	row, err := r.queries.GetExpenseByID(ctx, sqlc.GetExpenseByIDParams{
		ID:         expenseID,
		BusinessID: businessID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrExpenseNotFound
		}
		logger.Ctx(ctx).Error("Failed to get expense by ID", logger.Error(err))

		return nil, err
	}

	e := expenseFromSqlc(row)

	return &e, nil
}

func (r *Repository) CreateExpense(ctx context.Context, businessID int64, e domain.Expense) (int64, error) {
	result, err := r.queries.CreateExpense(ctx, sqlc.CreateExpenseParams{
		OwnerID:         businessID,
		BusinessID:      businessID,
		CategoryID:      sqlNullInt64(e.CategoryID),
		Description:     e.Description,
		Amount:          fmtFloat(e.Amount),
		Currency:        sqlNullString(e.Currency),
		ExpenseDate:     e.ExpenseDate,
		Vendor:          sqlNullString(e.Vendor),
		PaymentMethod:   sqlc.NullFinanceExpensesPaymentMethod{FinanceExpensesPaymentMethod: sqlc.FinanceExpensesPaymentMethod(e.PaymentMethod), Valid: e.PaymentMethod != ""},
		ReceiptUrl:      sqlNullString(e.ReceiptURL),
		IsTaxDeductible: sql.NullBool{Bool: e.IsTaxDeductible, Valid: true},
		TaxAmount:       sqlNullString(fmtFloat(e.TaxAmount)),
		Notes:           sqlNullString(e.Notes),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create expense", logger.Error(err))

		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to read created expense ID", logger.Error(err))

		return 0, err
	}

	return id, nil
}

func (r *Repository) UpdateExpense(ctx context.Context, businessID int64, e domain.Expense) error {
	err := r.queries.UpdateExpense(ctx, sqlc.UpdateExpenseParams{
		CategoryID:      sqlNullInt64(e.CategoryID),
		Description:     e.Description,
		Amount:          fmtFloat(e.Amount),
		Currency:        sqlNullString(e.Currency),
		ExpenseDate:     e.ExpenseDate,
		Vendor:          sqlNullString(e.Vendor),
		PaymentMethod:   sqlc.NullFinanceExpensesPaymentMethod{FinanceExpensesPaymentMethod: sqlc.FinanceExpensesPaymentMethod(e.PaymentMethod), Valid: e.PaymentMethod != ""},
		ReceiptUrl:      sqlNullString(e.ReceiptURL),
		IsTaxDeductible: sql.NullBool{Bool: e.IsTaxDeductible, Valid: true},
		TaxAmount:       sqlNullString(fmtFloat(e.TaxAmount)),
		Notes:           sqlNullString(e.Notes),
		ID:              e.ID,
		BusinessID:      businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update expense", logger.Error(err))
	}

	return err
}

func (r *Repository) DeleteExpense(ctx context.Context, businessID, expenseID int64) error {
	err := r.queries.DeleteExpense(ctx, sqlc.DeleteExpenseParams{
		ID:         expenseID,
		BusinessID: businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete expense", logger.Error(err))
	}

	return err
}

// ---------------------------------------------------------------------------
// Expense Categories
// ---------------------------------------------------------------------------

func (r *Repository) ListExpenseCategories(ctx context.Context, businessID int64) ([]domain.ExpenseCategory, error) {
	rows, err := r.queries.ListExpenseCategories(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list expense categories", logger.Error(err))

		return nil, err
	}

	categories := make([]domain.ExpenseCategory, 0, len(rows))
	for _, row := range rows {
		categories = append(categories, domain.ExpenseCategory{
			ID:         row.ID,
			OwnerID:    row.OwnerID,
			BusinessID: row.BusinessID,
			Name:       row.Name,
			Color:      nullStr(row.Color, ""),
			Icon:       nullStr(row.Icon, ""),
			CreatedAt:  nullTime(row.CreatedAt),
		})
	}

	return categories, nil
}

func (r *Repository) CreateExpenseCategory(ctx context.Context, businessID int64, c domain.ExpenseCategory) (int64, error) {
	result, err := r.queries.CreateExpenseCategory(ctx, sqlc.CreateExpenseCategoryParams{
		OwnerID:    businessID,
		BusinessID: businessID,
		Name:       c.Name,
		Color:      sqlNullString(c.Color),
		Icon:       sqlNullString(c.Icon),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create expense category", logger.Error(err))

		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to read created expense category ID", logger.Error(err))

		return 0, err
	}

	return id, nil
}

// ---------------------------------------------------------------------------
// Tax Config
// ---------------------------------------------------------------------------

func (r *Repository) ListTaxConfigs(ctx context.Context, businessID int64) ([]domain.TaxConfig, error) {
	rows, err := r.queries.ListTaxConfigs(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list tax configs", logger.Error(err))

		return nil, err
	}

	configs := make([]domain.TaxConfig, 0, len(rows))
	for _, row := range rows {
		configs = append(configs, domain.TaxConfig{
			ID:         row.ID,
			OwnerID:    row.OwnerID,
			BusinessID: row.BusinessID,
			Name:       row.Name,
			Rate:       parseFloat(row.Rate),
			Type:       nullEnumStr(row.Type.FinanceTaxConfigType, row.Type.Valid, "vat"),
			IsDefault:  nullBool(row.IsDefault),
			IsActive:   nullBool(row.IsActive),
			CreatedAt:  nullTime(row.CreatedAt),
		})
	}

	return configs, nil
}

func (r *Repository) CreateTaxConfig(ctx context.Context, businessID int64, t domain.TaxConfig) (int64, error) {
	result, err := r.queries.CreateTaxConfig(ctx, sqlc.CreateTaxConfigParams{
		OwnerID:    businessID,
		BusinessID: businessID,
		Name:       t.Name,
		Rate:       fmtFloat(t.Rate),
		Type:       sqlc.NullFinanceTaxConfigType{FinanceTaxConfigType: sqlc.FinanceTaxConfigType(t.Type), Valid: t.Type != ""},
		IsDefault:  sql.NullBool{Bool: t.IsDefault, Valid: true},
		IsActive:   sql.NullBool{Bool: true, Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create tax config", logger.Error(err))

		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to read created tax config ID", logger.Error(err))

		return 0, err
	}

	return id, nil
}

// ---------------------------------------------------------------------------
// Reports
// ---------------------------------------------------------------------------

func (r *Repository) SumPaidInvoices(ctx context.Context, businessID int64, dateFrom, dateTo string) (float64, error) {
	result, err := r.queries.SumPaidInvoices(ctx, sqlc.SumPaidInvoicesParams{
		BusinessID:   businessID,
		FromPaidDate: sqlNullTimeStr(dateFrom),
		ToPaidDate:   sqlNullTimeStr(dateTo),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to sum paid invoices", logger.Error(err))

		return 0, err
	}

	return toFloat64FromInterface(result), nil
}

func (r *Repository) SumExpenses(ctx context.Context, businessID int64, dateFrom, dateTo string) (float64, error) {
	dfTime, err := time.Parse("2006-01-02", dateFrom)
	if err != nil && dateFrom != "" {
		logger.Ctx(ctx).Error("Failed to parse dateFrom for SumExpenses", logger.Error(err))
		return 0, err
	}
	dtTime, err := time.Parse("2006-01-02", dateTo)
	if err != nil && dateTo != "" {
		logger.Ctx(ctx).Error("Failed to parse dateTo for SumExpenses", logger.Error(err))
		return 0, err
	}

	result, err := r.queries.SumExpenses(ctx, sqlc.SumExpensesParams{
		BusinessID:      businessID,
		FromExpenseDate: dfTime,
		ToExpenseDate:   dtTime,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to sum expenses", logger.Error(err))

		return 0, err
	}

	return toFloat64FromInterface(result), nil
}

func (r *Repository) SumTotalReceivable(ctx context.Context, businessID int64) (float64, error) {
	result, err := r.queries.SumTotalReceivable(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to sum total receivable", logger.Error(err))

		return 0, err
	}

	return toFloat64FromInterface(result), nil
}

func (r *Repository) SumOverdueAmount(ctx context.Context, businessID int64) (float64, error) {
	result, err := r.queries.SumOverdueAmount(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to sum overdue amount", logger.Error(err))

		return 0, err
	}

	return toFloat64FromInterface(result), nil
}

func (r *Repository) SumTaxCollected(ctx context.Context, businessID int64, dateFrom, dateTo string) (float64, error) {
	result, err := r.queries.SumTaxCollected(ctx, sqlc.SumTaxCollectedParams{
		BusinessID:   businessID,
		FromPaidDate: sqlNullTimeStr(dateFrom),
		ToPaidDate:   sqlNullTimeStr(dateTo),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to sum tax collected", logger.Error(err))

		return 0, err
	}

	return toFloat64FromInterface(result), nil
}

func (r *Repository) SumTaxOwed(ctx context.Context, businessID int64, dateFrom, dateTo string) (float64, error) {
	dfTime, err := time.Parse("2006-01-02", dateFrom)
	if err != nil && dateFrom != "" {
		logger.Ctx(ctx).Error("Failed to parse dateFrom for SumTaxOwed", logger.Error(err))
		return 0, err
	}
	dtTime, err := time.Parse("2006-01-02", dateTo)
	if err != nil && dateTo != "" {
		logger.Ctx(ctx).Error("Failed to parse dateTo for SumTaxOwed", logger.Error(err))
		return 0, err
	}

	result, err := r.queries.SumTaxOwed(ctx, sqlc.SumTaxOwedParams{
		BusinessID:      businessID,
		FromExpenseDate: dfTime,
		ToExpenseDate:   dtTime,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to sum tax owed", logger.Error(err))

		return 0, err
	}

	return toFloat64FromInterface(result), nil
}

// CountInvoicesByStatus returns how many invoices the owner has in each
// status bucket. The returned map uses status strings as keys (e.g. "draft",
// "paid"). An empty map is returned when the owner has no invoices.
func (r *Repository) CountInvoicesByStatus(ctx context.Context, businessID int64) (map[string]int64, error) {
	rows, err := r.queries.CountInvoicesByStatus(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to count invoices by status", logger.Error(err))

		return nil, err
	}

	counts := make(map[string]int64, len(rows))
	for _, row := range rows {
		counts[string(row.Status.FinanceInvoicesStatus)] = row.Cnt
	}

	return counts, nil
}

// GetOverdueInvoices returns one page of unpaid invoices whose due_date is
// before the given cutoff (expressed as YYYY-MM-DD). Pass an empty string to
// use today's date.
func (r *Repository) GetOverdueInvoices(ctx context.Context, businessID int64, beforeDate string, limit, offset int32) ([]domain.Invoice, error) {
	before, err := parseOverdueCutoff(ctx, beforeDate)
	if err != nil {
		return nil, err
	}

	rows, err := r.queries.ListOverdueInvoices(ctx, sqlc.ListOverdueInvoicesParams{
		BusinessID: businessID,
		DueDate:    before,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list overdue invoices", logger.Error(err))

		return nil, err
	}

	invoices := make([]domain.Invoice, 0, len(rows))
	for _, row := range rows {
		invoices = append(invoices, invoiceFromSqlc(row))
	}

	return invoices, nil
}

// CountOverdueInvoices returns the total number of unpaid invoices whose
// due_date is before the given cutoff (expressed as YYYY-MM-DD). Pass an
// empty string to use today's date.
func (r *Repository) CountOverdueInvoices(ctx context.Context, businessID int64, beforeDate string) (int64, error) {
	before, err := parseOverdueCutoff(ctx, beforeDate)
	if err != nil {
		return 0, err
	}

	count, err := r.queries.CountOverdueInvoices(ctx, sqlc.CountOverdueInvoicesParams{
		BusinessID: businessID,
		DueDate:    before,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to count overdue invoices", logger.Error(err))

		return 0, err
	}

	return count, nil
}

func parseOverdueCutoff(ctx context.Context, beforeDate string) (time.Time, error) {
	if beforeDate == "" {
		beforeDate = time.Now().Format("2006-01-02")
	}

	before, err := time.Parse("2006-01-02", beforeDate)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to parse beforeDate for overdue invoices", logger.Error(err))

		return time.Time{}, err
	}

	return before, nil
}

// ---------------------------------------------------------------------------
// Mapping helpers
// ---------------------------------------------------------------------------

func invoiceFromSqlc(row sqlc.FinanceInvoice) domain.Invoice {
	inv := domain.Invoice{
		ID:             row.ID,
		OwnerID:        row.OwnerID,
		BusinessID:     row.BusinessID,
		InvoiceNumber:  row.InvoiceNumber,
		ClientName:     row.ClientName,
		ClientEmail:    nullStr(row.ClientEmail, ""),
		ClientAddress:  nullStr(row.ClientAddress, ""),
		ClientTaxID:    nullStr(row.ClientTaxID, ""),
		CustomerID:     nullableInt64Ptr(row.CustomerID),
		Status:         string(row.Status.FinanceInvoicesStatus),
		IssueDate:      row.IssueDate,
		DueDate:        row.DueDate,
		Currency:       nullStr(row.Currency, "USD"),
		Subtotal:       parseFloatNull(row.Subtotal),
		DiscountType:   nullEnumStr(row.DiscountType.FinanceInvoicesDiscountType, row.DiscountType.Valid, "fixed"),
		DiscountValue:  parseFloatNull(row.DiscountValue),
		DiscountAmount: parseFloatNull(row.DiscountAmount),
		TaxRate:        parseFloatNull(row.TaxRate),
		TaxAmount:      parseFloatNull(row.TaxAmount),
		Total:          parseFloatNull(row.Total),
		AmountPaid:     parseFloatNull(row.AmountPaid),
		AmountDue:      parseFloatNull(row.AmountDue),
		Notes:          nullStr(row.Notes, ""),
		Terms:          nullStr(row.Terms, ""),
		CreatedAt:      nullTime(row.CreatedAt),
		UpdatedAt:      nullTime(row.UpdatedAt),
	}

	if row.PaidDate.Valid {
		t := row.PaidDate.Time
		inv.PaidDate = &t
	}

	return inv
}

func expenseFromSqlc(row sqlc.FinanceExpense) domain.Expense {
	return domain.Expense{
		ID:              row.ID,
		OwnerID:         row.OwnerID,
		BusinessID:      row.BusinessID,
		CategoryID:      toInt64(row.CategoryID),
		Description:     row.Description,
		Amount:          parseFloat(row.Amount),
		Currency:        nullStr(row.Currency, "USD"),
		ExpenseDate:     row.ExpenseDate,
		Vendor:          nullStr(row.Vendor, ""),
		PaymentMethod:   nullEnumStr(row.PaymentMethod.FinanceExpensesPaymentMethod, row.PaymentMethod.Valid, "cash"),
		ReceiptURL:      nullStr(row.ReceiptUrl, ""),
		IsTaxDeductible: nullBool(row.IsTaxDeductible),
		TaxAmount:       parseFloatNull(row.TaxAmount),
		Notes:           nullStr(row.Notes, ""),
		Status:          nullStr(row.Status, "pending"),
		CreatedAt:       nullTime(row.CreatedAt),
		UpdatedAt:       nullTime(row.UpdatedAt),
	}
}

// ---------------------------------------------------------------------------
// Conversion helpers
// ---------------------------------------------------------------------------

func sqlNullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

func sqlNullInt64(i int64) sql.NullInt64 {
	return sql.NullInt64{Int64: i, Valid: i != 0}
}

func fmtFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', 2, 64)
}

// paymentAmountToCents rounds an API-supplied payment amount to whole cents.
func paymentAmountToCents(amount float64) money.Amount {
	return money.Amount(int64(math.Round(amount * 100)))
}

// nullDecimalToCents parses a DECIMAL column read back as a nullable string
// into integer cents via the shared money parser, so no float64 rounding
// error is introduced on the way out of the database. Unparseable values
// degrade to zero with a warning rather than failing the payment write.
func nullDecimalToCents(value sql.NullString) money.Amount {
	if !value.Valid || strings.TrimSpace(value.String) == "" {
		return 0
	}
	cents, err := money.FromString(value.String)
	if err != nil {
		logger.Global().Warn("failed to parse invoice decimal value",
			logger.String("value", value.String),
			logger.Error(err),
		)
		return 0
	}
	return cents
}

func parseFloat(s string) float64 {
	if s == "" {
		return 0
	}

	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		logger.Global().Warn("failed to parse float from database",
			logger.String("value", s),
			logger.Error(err),
		)

		return 0
	}

	return v
}

func parseFloatNull(s sql.NullString) float64 {
	if !s.Valid || s.String == "" {
		return 0
	}

	v, err := strconv.ParseFloat(s.String, 64)
	if err != nil {
		logger.Global().Warn("failed to parse nullable float from database",
			logger.String("value", s.String),
			logger.Error(err),
		)

		return 0
	}

	return v
}

func nullStr(s sql.NullString, def string) string {
	if s.Valid {
		return s.String
	}

	return def
}

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

func toInt64(i sql.NullInt64) int64 {
	if i.Valid {
		return i.Int64
	}

	return 0
}

func toInt32(i sql.NullInt32) int32 {
	if i.Valid {
		return i.Int32
	}

	return 0
}

func nullEnumStr[T ~string](val T, valid bool, def string) string {
	if valid {
		return string(val)
	}

	return def
}

func sqlNullTimeStr(s string) sql.NullTime {
	if s == "" {
		return sql.NullTime{}
	}

	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return sql.NullTime{}
	}

	return sql.NullTime{Time: t, Valid: true}
}

func toFloat64FromInterface(v interface{}) float64 {
	if v == nil {
		return 0
	}

	var str string

	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int64:
		return float64(val)
	case []byte:
		str = string(val)
	case string:
		str = val
	default:
		str = fmt.Sprintf("%v", val)
	}

	if str == "" {
		return 0
	}

	f, err := strconv.ParseFloat(str, 64)
	if err != nil {
		logger.Global().Warn("failed to parse float from interface value",
			logger.String("value", str),
			logger.Error(err),
		)

		return 0
	}

	return f
}

func nullableInt64Ptr(v sql.NullInt64) *int64 {
	if !v.Valid {
		return nil
	}
	out := v.Int64
	return &out
}

func nullableInt64FromPtr(v *int64) sql.NullInt64 {
	if v == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *v, Valid: true}
}

// ListInvoicesByCustomer pages one customer's invoices off the
// business+customer index.
func (r *Repository) ListInvoicesByCustomer(ctx context.Context, businessID, customerID int64, limit, offset int) ([]domain.Invoice, error) {
	rows, err := r.queries.ListInvoicesByCustomer(ctx, sqlc.ListInvoicesByCustomerParams{
		BusinessID: businessID,
		CustomerID: sql.NullInt64{Int64: customerID, Valid: true},
		Limit:      int32(limit),  // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		Offset:     int32(offset), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to list invoices by customer", logger.Any("business_id", businessID), logger.Any("customer_id", customerID), logger.Error(err))
		return nil, err
	}
	out := make([]domain.Invoice, 0, len(rows))
	for _, row := range rows {
		out = append(out, invoiceFromSqlc(row))
	}
	return out, nil
}

// CountInvoicesByCustomer reports the customer's invoice count.
func (r *Repository) CountInvoicesByCustomer(ctx context.Context, businessID, customerID int64) (int64, error) {
	return r.queries.CountInvoicesByCustomer(ctx, sqlc.CountInvoicesByCustomerParams{
		BusinessID: businessID,
		CustomerID: sql.NullInt64{Int64: customerID, Valid: true},
	})
}
