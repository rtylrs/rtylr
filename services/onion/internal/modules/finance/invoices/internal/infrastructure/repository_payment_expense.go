package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/Masterminds/squirrel"

	"github.com/voxire/onion/internal/modules/finance/invoices/internal/domain"
	"github.com/voxire/onion/internal/modules/finance/invoices/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/money"
	"github.com/voxire/onion/internal/platform/searchutil"
)

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

	// Compare in integer cents while the invoice row is locked. This prevents
	// concurrent payments from each appearing valid against the same balance.
	paymentCents := paymentAmountToCents(input.Amount)
	totalCents := nullDecimalToCents(row.Total)
	paidCents := nullDecimalToCents(row.AmountPaid)
	if paymentCents > totalCents-paidCents {
		return nil, domain.ErrPaymentExceedsBalance
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

	// Accumulate in integer cents: float64 residue from parsed DECIMAL strings
	// could otherwise leave newAmountDue at +2.4e-15 and keep a fully paid
	// invoice in "partial" forever (or flip one early).
	newPaidCents := paidCents + paymentCents
	newDueCents := totalCents - newPaidCents
	if newDueCents < 0 {
		newDueCents = 0
	}
	newStatus, paidDate := paymentTotalsStatus(newDueCents, p.PaymentDate)
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

// paymentTotalsStatus is the only place that decides the money-derived invoice
// statuses, and it is reached only by a recorded payment: "partial" while a
// balance is still outstanding, "paid" (with the paid date that revenue
// reporting is bucketed by) once nothing is. Compare in whole cents, so float
// residue can neither keep a settled invoice in "partial" nor flip an unpaid
// one early.
func paymentTotalsStatus(dueCents money.Amount, paymentDate time.Time) (string, *time.Time) {
	if dueCents > 0 {
		return domain.StatusPartial, nil
	}
	paidDate := paymentDate
	return domain.StatusPaid, &paidDate
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
		e, scanErr := expenseFromRow(rows)
		if scanErr != nil {
			logger.Ctx(ctx).Error("Failed to scan expense row", logger.Error(scanErr))

			return nil, scanErr
		}

		expenses = append(expenses, e)
	}

	return expenses, rows.Err()
}
