package infrastructure

import (
	"context"
	"database/sql"
	"time"

	"github.com/voxire/onion/internal/modules/invoices/internal/domain"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{
		pool: dbclient.Pool(),
	}
}

// ---------------------------------------------------------------------------
// Invoices
// ---------------------------------------------------------------------------

func (r *Repository) ListInvoices(ctx context.Context, ownerID int64, status, search, dateFrom, dateTo string) ([]domain.Invoice, error) {
	query := `
		SELECT id, owner_id, invoice_number, client_name, client_email, client_address, client_tax_id,
		       status, issue_date, due_date, paid_date, currency, subtotal, discount_type,
		       discount_value, discount_amount, tax_rate, tax_amount, total, amount_paid, amount_due,
		       notes, terms, created_at, updated_at
		FROM finance_invoices
		WHERE owner_id = ?`

	args := []interface{}{ownerID}

	if status != "" {
		query += " AND status = ?"
		args = append(args, status)
	}

	if search != "" {
		query += " AND (client_name LIKE ? OR invoice_number LIKE ?)"
		like := "%" + search + "%"
		args = append(args, like, like)
	}

	if dateFrom != "" {
		query += " AND issue_date >= ?"
		args = append(args, dateFrom)
	}

	if dateTo != "" {
		query += " AND issue_date <= ?"
		args = append(args, dateTo)
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list invoices", logger.Error(err))

		return nil, err
	}
	defer rows.Close()

	var invoices []domain.Invoice

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

func (r *Repository) GetInvoiceByID(ctx context.Context, ownerID, invoiceID int64) (*domain.Invoice, error) {
	query := `
		SELECT id, owner_id, invoice_number, client_name, client_email, client_address, client_tax_id,
		       status, issue_date, due_date, paid_date, currency, subtotal, discount_type,
		       discount_value, discount_amount, tax_rate, tax_amount, total, amount_paid, amount_due,
		       notes, terms, created_at, updated_at
		FROM finance_invoices
		WHERE id = ? AND owner_id = ?`

	rows, err := r.pool.QueryContext(ctx, query, invoiceID, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get invoice by ID", logger.Error(err))

		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, domain.ErrInvoiceNotFound
	}

	var inv domain.Invoice
	if err := inv.FromRow(rows); err != nil {
		logger.Ctx(ctx).Error("Failed to scan invoice", logger.Error(err))

		return nil, err
	}

	return &inv, rows.Err()
}

func (r *Repository) CreateInvoice(ctx context.Context, ownerID int64, inv domain.Invoice) (int64, error) {
	query := `
		INSERT INTO finance_invoices
		  (owner_id, invoice_number, client_name, client_email, client_address, client_tax_id,
		   status, issue_date, due_date, currency, subtotal, discount_type, discount_value,
		   discount_amount, tax_rate, tax_amount, total, amount_paid, amount_due, notes, terms)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`

	result, err := r.pool.ExecContext(ctx, query,
		ownerID,
		inv.InvoiceNumber,
		inv.ClientName,
		sqlNullString(inv.ClientEmail),
		sqlNullString(inv.ClientAddress),
		sqlNullString(inv.ClientTaxID),
		inv.Status,
		inv.IssueDate.Format("2006-01-02"),
		inv.DueDate.Format("2006-01-02"),
		inv.Currency,
		inv.Subtotal,
		inv.DiscountType,
		inv.DiscountValue,
		inv.DiscountAmount,
		inv.TaxRate,
		inv.TaxAmount,
		inv.Total,
		inv.AmountPaid,
		inv.AmountDue,
		sqlNullString(inv.Notes),
		sqlNullString(inv.Terms),
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create invoice", logger.Error(err))

		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to read created invoice ID", logger.Error(err))

		return 0, err
	}

	return id, nil
}

func (r *Repository) CreateInvoiceItems(ctx context.Context, invoiceID int64, items []domain.InvoiceItem) error {
	query := `
		INSERT INTO finance_invoice_items
		  (invoice_id, description, quantity, unit, unit_price, tax_rate, discount_pct, total, sort_order)
		VALUES (?,?,?,?,?,?,?,?,?)`

	for _, item := range items {
		_, err := r.pool.ExecContext(ctx, query,
			invoiceID,
			item.Description,
			item.Quantity,
			sqlNullString(item.Unit),
			item.UnitPrice,
			item.TaxRate,
			item.DiscountPct,
			item.Total,
			item.SortOrder,
		)
		if err != nil {
			logger.Ctx(ctx).Error("Failed to create invoice item", logger.Error(err))

			return err
		}
	}

	return nil
}

func (r *Repository) GetInvoiceItems(ctx context.Context, invoiceID int64) ([]domain.InvoiceItem, error) {
	query := `
		SELECT id, invoice_id, description, quantity, unit, unit_price, tax_rate, discount_pct, total, sort_order
		FROM finance_invoice_items
		WHERE invoice_id = ?
		ORDER BY sort_order ASC, id ASC`

	rows, err := r.pool.QueryContext(ctx, query, invoiceID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get invoice items", logger.Error(err))

		return nil, err
	}
	defer rows.Close()

	var items []domain.InvoiceItem

	for rows.Next() {
		var item domain.InvoiceItem
		if scanErr := item.FromRow(rows); scanErr != nil {
			logger.Ctx(ctx).Error("Failed to scan invoice item row", logger.Error(scanErr))

			return nil, scanErr
		}

		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *Repository) DeleteInvoiceItems(ctx context.Context, invoiceID int64) error {
	_, err := r.pool.ExecContext(ctx, "DELETE FROM finance_invoice_items WHERE invoice_id = ?", invoiceID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete invoice items", logger.Error(err))
	}

	return err
}

func (r *Repository) UpdateInvoice(ctx context.Context, ownerID int64, inv domain.Invoice) error {
	query := `
		UPDATE finance_invoices SET
		  client_name = ?, client_email = ?, client_address = ?, client_tax_id = ?,
		  issue_date = ?, due_date = ?, currency = ?, subtotal = ?, discount_type = ?,
		  discount_value = ?, discount_amount = ?, tax_rate = ?, tax_amount = ?,
		  total = ?, amount_due = ?, notes = ?, terms = ?
		WHERE id = ? AND owner_id = ?`

	_, err := r.pool.ExecContext(ctx, query,
		inv.ClientName,
		sqlNullString(inv.ClientEmail),
		sqlNullString(inv.ClientAddress),
		sqlNullString(inv.ClientTaxID),
		inv.IssueDate.Format("2006-01-02"),
		inv.DueDate.Format("2006-01-02"),
		inv.Currency,
		inv.Subtotal,
		inv.DiscountType,
		inv.DiscountValue,
		inv.DiscountAmount,
		inv.TaxRate,
		inv.TaxAmount,
		inv.Total,
		inv.AmountDue,
		sqlNullString(inv.Notes),
		sqlNullString(inv.Terms),
		inv.ID,
		ownerID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update invoice", logger.Error(err))
	}

	return err
}

func (r *Repository) UpdateInvoiceStatus(ctx context.Context, ownerID, invoiceID int64, status string) error {
	_, err := r.pool.ExecContext(ctx,
		"UPDATE finance_invoices SET status = ? WHERE id = ? AND owner_id = ?",
		status, invoiceID, ownerID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update invoice status", logger.Error(err))
	}

	return err
}

func (r *Repository) UpdateInvoicePaymentTotals(ctx context.Context, invoiceID int64, amountPaid, amountDue float64, status string, paidDate *time.Time) error {
	if paidDate != nil {
		_, err := r.pool.ExecContext(ctx,
			"UPDATE finance_invoices SET amount_paid = ?, amount_due = ?, status = ?, paid_date = ? WHERE id = ?",
			amountPaid, amountDue, status, paidDate.Format("2006-01-02"), invoiceID,
		)
		if err != nil {
			logger.Ctx(ctx).Error("Failed to update invoice payment totals", logger.Error(err))
		}

		return err
	}

	_, err := r.pool.ExecContext(ctx,
		"UPDATE finance_invoices SET amount_paid = ?, amount_due = ?, status = ? WHERE id = ?",
		amountPaid, amountDue, status, invoiceID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update invoice payment totals", logger.Error(err))
	}

	return err
}

func (r *Repository) DeleteInvoice(ctx context.Context, ownerID, invoiceID int64) error {
	_, err := r.pool.ExecContext(ctx,
		"DELETE FROM finance_invoices WHERE id = ? AND owner_id = ? AND status = 'draft'",
		invoiceID, ownerID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete invoice", logger.Error(err))
	}

	return err
}

func (r *Repository) CountInvoiceSequence(ctx context.Context, ownerID int64, year int) (int64, error) {
	var count int64

	err := r.pool.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM finance_invoices WHERE owner_id = ? AND YEAR(created_at) = ?",
		ownerID, year,
	).Scan(&count)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to count invoice sequence", logger.Error(err))

		return 0, err
	}

	return count, nil
}

// ---------------------------------------------------------------------------
// Payments
// ---------------------------------------------------------------------------

func (r *Repository) CreatePayment(ctx context.Context, ownerID int64, p domain.Payment) (int64, error) {
	query := `
		INSERT INTO finance_payments
		  (owner_id, invoice_id, amount, currency, payment_date, method, reference, notes)
		VALUES (?,?,?,?,?,?,?,?)`

	result, err := r.pool.ExecContext(ctx, query,
		ownerID,
		p.InvoiceID,
		p.Amount,
		p.Currency,
		p.PaymentDate.Format("2006-01-02"),
		p.Method,
		sqlNullString(p.Reference),
		sqlNullString(p.Notes),
	)
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

func (r *Repository) GetPaymentsByInvoice(ctx context.Context, invoiceID int64) ([]domain.Payment, error) {
	query := `
		SELECT id, owner_id, invoice_id, amount, currency, payment_date, method, reference, notes, created_at
		FROM finance_payments
		WHERE invoice_id = ?
		ORDER BY payment_date ASC`

	rows, err := r.pool.QueryContext(ctx, query, invoiceID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get payments by invoice", logger.Error(err))

		return nil, err
	}
	defer rows.Close()

	var payments []domain.Payment

	for rows.Next() {
		var p domain.Payment
		if scanErr := p.FromRow(rows); scanErr != nil {
			logger.Ctx(ctx).Error("Failed to scan payment row", logger.Error(scanErr))

			return nil, scanErr
		}

		payments = append(payments, p)
	}

	return payments, rows.Err()
}

// ---------------------------------------------------------------------------
// Expenses
// ---------------------------------------------------------------------------

func (r *Repository) ListExpenses(ctx context.Context, ownerID int64, categoryID int64, search, dateFrom, dateTo string) ([]domain.Expense, error) {
	query := `
		SELECT id, owner_id, category_id, description, amount, currency, expense_date, vendor,
		       payment_method, receipt_url, is_tax_deductible, tax_amount, notes, created_at, updated_at
		FROM finance_expenses
		WHERE owner_id = ?`

	args := []interface{}{ownerID}

	if categoryID > 0 {
		query += " AND category_id = ?"
		args = append(args, categoryID)
	}

	if search != "" {
		query += " AND (description LIKE ? OR vendor LIKE ?)"
		like := "%" + search + "%"
		args = append(args, like, like)
	}

	if dateFrom != "" {
		query += " AND expense_date >= ?"
		args = append(args, dateFrom)
	}

	if dateTo != "" {
		query += " AND expense_date <= ?"
		args = append(args, dateTo)
	}

	query += " ORDER BY expense_date DESC"

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list expenses", logger.Error(err))

		return nil, err
	}
	defer rows.Close()

	var expenses []domain.Expense

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

func (r *Repository) GetExpenseByID(ctx context.Context, ownerID, expenseID int64) (*domain.Expense, error) {
	query := `
		SELECT id, owner_id, category_id, description, amount, currency, expense_date, vendor,
		       payment_method, receipt_url, is_tax_deductible, tax_amount, notes, created_at, updated_at
		FROM finance_expenses
		WHERE id = ? AND owner_id = ?`

	rows, err := r.pool.QueryContext(ctx, query, expenseID, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get expense by ID", logger.Error(err))

		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, domain.ErrExpenseNotFound
	}

	var e domain.Expense
	if err := e.FromRow(rows); err != nil {
		logger.Ctx(ctx).Error("Failed to scan expense", logger.Error(err))

		return nil, err
	}

	return &e, rows.Err()
}

func (r *Repository) CreateExpense(ctx context.Context, ownerID int64, e domain.Expense) (int64, error) {
	query := `
		INSERT INTO finance_expenses
		  (owner_id, category_id, description, amount, currency, expense_date, vendor,
		   payment_method, receipt_url, is_tax_deductible, tax_amount, notes)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`

	result, err := r.pool.ExecContext(ctx, query,
		ownerID,
		sqlNullInt64(e.CategoryID),
		e.Description,
		e.Amount,
		e.Currency,
		e.ExpenseDate.Format("2006-01-02"),
		sqlNullString(e.Vendor),
		e.PaymentMethod,
		sqlNullString(e.ReceiptURL),
		e.IsTaxDeductible,
		e.TaxAmount,
		sqlNullString(e.Notes),
	)
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

func (r *Repository) UpdateExpense(ctx context.Context, ownerID int64, e domain.Expense) error {
	query := `
		UPDATE finance_expenses SET
		  category_id = ?, description = ?, amount = ?, currency = ?, expense_date = ?,
		  vendor = ?, payment_method = ?, receipt_url = ?, is_tax_deductible = ?,
		  tax_amount = ?, notes = ?
		WHERE id = ? AND owner_id = ?`

	_, err := r.pool.ExecContext(ctx, query,
		sqlNullInt64(e.CategoryID),
		e.Description,
		e.Amount,
		e.Currency,
		e.ExpenseDate.Format("2006-01-02"),
		sqlNullString(e.Vendor),
		e.PaymentMethod,
		sqlNullString(e.ReceiptURL),
		e.IsTaxDeductible,
		e.TaxAmount,
		sqlNullString(e.Notes),
		e.ID,
		ownerID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update expense", logger.Error(err))
	}

	return err
}

func (r *Repository) DeleteExpense(ctx context.Context, ownerID, expenseID int64) error {
	_, err := r.pool.ExecContext(ctx,
		"DELETE FROM finance_expenses WHERE id = ? AND owner_id = ?",
		expenseID, ownerID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete expense", logger.Error(err))
	}

	return err
}

// ---------------------------------------------------------------------------
// Expense Categories
// ---------------------------------------------------------------------------

func (r *Repository) ListExpenseCategories(ctx context.Context, ownerID int64) ([]domain.ExpenseCategory, error) {
	query := `
		SELECT id, owner_id, name, color, icon, created_at
		FROM finance_expense_categories
		WHERE owner_id = ?
		ORDER BY name ASC`

	rows, err := r.pool.QueryContext(ctx, query, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list expense categories", logger.Error(err))

		return nil, err
	}
	defer rows.Close()

	var categories []domain.ExpenseCategory

	for rows.Next() {
		var c domain.ExpenseCategory
		if scanErr := c.FromRow(rows); scanErr != nil {
			logger.Ctx(ctx).Error("Failed to scan expense category row", logger.Error(scanErr))

			return nil, scanErr
		}

		categories = append(categories, c)
	}

	return categories, rows.Err()
}

func (r *Repository) CreateExpenseCategory(ctx context.Context, ownerID int64, c domain.ExpenseCategory) (int64, error) {
	query := `
		INSERT INTO finance_expense_categories (owner_id, name, color, icon)
		VALUES (?,?,?,?)`

	result, err := r.pool.ExecContext(ctx, query,
		ownerID,
		c.Name,
		sqlNullString(c.Color),
		sqlNullString(c.Icon),
	)
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

func (r *Repository) ListTaxConfigs(ctx context.Context, ownerID int64) ([]domain.TaxConfig, error) {
	query := `
		SELECT id, owner_id, name, rate, type, is_default, is_active, created_at
		FROM finance_tax_config
		WHERE owner_id = ?
		ORDER BY is_default DESC, name ASC`

	rows, err := r.pool.QueryContext(ctx, query, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list tax configs", logger.Error(err))

		return nil, err
	}
	defer rows.Close()

	var configs []domain.TaxConfig

	for rows.Next() {
		var t domain.TaxConfig
		if scanErr := t.FromRow(rows); scanErr != nil {
			logger.Ctx(ctx).Error("Failed to scan tax config row", logger.Error(scanErr))

			return nil, scanErr
		}

		configs = append(configs, t)
	}

	return configs, rows.Err()
}

func (r *Repository) CreateTaxConfig(ctx context.Context, ownerID int64, t domain.TaxConfig) (int64, error) {
	query := `
		INSERT INTO finance_tax_config (owner_id, name, rate, type, is_default, is_active)
		VALUES (?,?,?,?,?,?)`

	result, err := r.pool.ExecContext(ctx, query,
		ownerID,
		t.Name,
		t.Rate,
		t.Type,
		t.IsDefault,
		true,
	)
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

func (r *Repository) SumPaidInvoices(ctx context.Context, ownerID int64, dateFrom, dateTo string) (float64, error) {
	var total sql.NullFloat64

	err := r.pool.QueryRowContext(ctx,
		"SELECT COALESCE(SUM(total),0) FROM finance_invoices WHERE owner_id = ? AND status = 'paid' AND paid_date BETWEEN ? AND ?",
		ownerID, dateFrom, dateTo,
	).Scan(&total)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to sum paid invoices", logger.Error(err))

		return 0, err
	}

	return total.Float64, nil
}

func (r *Repository) SumExpenses(ctx context.Context, ownerID int64, dateFrom, dateTo string) (float64, error) {
	var total sql.NullFloat64

	err := r.pool.QueryRowContext(ctx,
		"SELECT COALESCE(SUM(amount),0) FROM finance_expenses WHERE owner_id = ? AND expense_date BETWEEN ? AND ?",
		ownerID, dateFrom, dateTo,
	).Scan(&total)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to sum expenses", logger.Error(err))

		return 0, err
	}

	return total.Float64, nil
}

func (r *Repository) SumTotalReceivable(ctx context.Context, ownerID int64) (float64, error) {
	var total sql.NullFloat64

	err := r.pool.QueryRowContext(ctx,
		"SELECT COALESCE(SUM(amount_due),0) FROM finance_invoices WHERE owner_id = ? AND status IN ('sent','viewed','partial','overdue')",
		ownerID,
	).Scan(&total)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to sum total receivable", logger.Error(err))

		return 0, err
	}

	return total.Float64, nil
}

func (r *Repository) SumOverdueAmount(ctx context.Context, ownerID int64) (float64, error) {
	var total sql.NullFloat64

	err := r.pool.QueryRowContext(ctx,
		"SELECT COALESCE(SUM(amount_due),0) FROM finance_invoices WHERE owner_id = ? AND status = 'overdue'",
		ownerID,
	).Scan(&total)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to sum overdue amount", logger.Error(err))

		return 0, err
	}

	return total.Float64, nil
}

func (r *Repository) SumTaxCollected(ctx context.Context, ownerID int64, dateFrom, dateTo string) (float64, error) {
	var total sql.NullFloat64

	err := r.pool.QueryRowContext(ctx,
		"SELECT COALESCE(SUM(tax_amount),0) FROM finance_invoices WHERE owner_id = ? AND status = 'paid' AND paid_date BETWEEN ? AND ?",
		ownerID, dateFrom, dateTo,
	).Scan(&total)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to sum tax collected", logger.Error(err))

		return 0, err
	}

	return total.Float64, nil
}

func (r *Repository) SumTaxOwed(ctx context.Context, ownerID int64, dateFrom, dateTo string) (float64, error) {
	var total sql.NullFloat64

	err := r.pool.QueryRowContext(ctx,
		"SELECT COALESCE(SUM(tax_amount),0) FROM finance_expenses WHERE owner_id = ? AND is_tax_deductible = TRUE AND expense_date BETWEEN ? AND ?",
		ownerID, dateFrom, dateTo,
	).Scan(&total)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to sum tax owed", logger.Error(err))

		return 0, err
	}

	return total.Float64, nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func sqlNullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

func sqlNullInt64(i int64) sql.NullInt64 {
	return sql.NullInt64{Int64: i, Valid: i != 0}
}
