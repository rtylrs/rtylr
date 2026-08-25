-- name: GetInvoiceByID :one
SELECT id, owner_id, business_id, invoice_number, client_name, client_email, client_address, client_tax_id, customer_id,
       status, issue_date, due_date, paid_date, reporting_date, currency, subtotal, discount_type,
       discount_value, discount_amount, tax_rate, tax_amount, total, amount_paid, amount_due,
       notes, terms, created_at, updated_at
FROM finance_invoices
WHERE id = ? AND business_id = ?;

-- name: GetInvoiceStatusForUpdate :one
SELECT status
FROM finance_invoices
WHERE id = ? AND business_id = ?
FOR UPDATE;

-- name: GetInvoiceByIDForUpdate :one
SELECT id, owner_id, business_id, invoice_number, client_name, client_email, client_address, client_tax_id, customer_id,
       status, issue_date, due_date, paid_date, reporting_date, currency, subtotal, discount_type,
       discount_value, discount_amount, tax_rate, tax_amount, total, amount_paid, amount_due,
       notes, terms, created_at, updated_at
FROM finance_invoices
WHERE id = ? AND business_id = ?
FOR UPDATE;

-- name: ListInvoicesByOwner :many
SELECT id, owner_id, business_id, invoice_number, client_name, client_email, client_address, client_tax_id, customer_id,
       status, issue_date, due_date, paid_date, reporting_date, currency, subtotal, discount_type,
       discount_value, discount_amount, tax_rate, tax_amount, total, amount_paid, amount_due,
       notes, terms, created_at, updated_at
FROM finance_invoices
WHERE business_id = ?
ORDER BY created_at DESC, id DESC
LIMIT ? OFFSET ?;

-- name: CountInvoicesByOwner :one
SELECT COUNT(*) FROM finance_invoices WHERE business_id = ?;

-- name: ListInvoicesByOwnerAndStatus :many
SELECT id, owner_id, business_id, invoice_number, client_name, client_email, client_address, client_tax_id, customer_id,
       status, issue_date, due_date, paid_date, reporting_date, currency, subtotal, discount_type,
       discount_value, discount_amount, tax_rate, tax_amount, total, amount_paid, amount_due,
       notes, terms, created_at, updated_at
FROM finance_invoices
WHERE business_id = ? AND status = ?
ORDER BY created_at DESC, id DESC
LIMIT 500;

-- name: CreateInvoice :execresult
INSERT INTO finance_invoices
  (owner_id, business_id, invoice_number, client_name, client_email, client_address, client_tax_id, customer_id,
   status, issue_date, due_date, currency, subtotal, discount_type, discount_value,
   discount_amount, tax_rate, tax_amount, total, amount_paid, amount_due, notes, terms)
VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?);

-- name: UpdateInvoice :exec
UPDATE finance_invoices SET
  client_name = ?, client_email = ?, client_address = ?, client_tax_id = ?, customer_id = ?,
  issue_date = ?, due_date = ?, currency = ?, subtotal = ?, discount_type = ?,
  discount_value = ?, discount_amount = ?, tax_rate = ?, tax_amount = ?,
  total = ?, amount_due = ?, notes = ?, terms = ?
WHERE id = ? AND business_id = ? AND status = 'draft';

-- name: UpdateInvoiceStatus :execresult
UPDATE finance_invoices
SET status = sqlc.arg(status)
WHERE id = ? AND business_id = ?
  AND status NOT IN ('paid', 'void', 'cancelled')
  AND (sqlc.arg(status) <> 'draft' OR status = 'draft');

-- name: UpdateInvoicePaymentTotalsWithPaidDate :exec
UPDATE finance_invoices SET amount_paid = ?, amount_due = ?, status = ?, paid_date = ? WHERE id = ? AND business_id = ?;

-- name: UpdateInvoicePaymentTotals :exec
UPDATE finance_invoices SET amount_paid = ?, amount_due = ?, status = ? WHERE id = ? AND business_id = ?;

-- name: DeleteInvoice :exec
DELETE FROM finance_invoices WHERE id = ? AND business_id = ? AND status = 'draft';

-- name: CountInvoiceSequence :one
SELECT COUNT(*) FROM finance_invoices
WHERE business_id = ?
  AND created_at >= sqlc.arg(created_at_start)
  AND created_at < sqlc.arg(created_at_end);

-- name: CountInvoicesByStatus :many
SELECT status, COUNT(*) AS cnt
FROM finance_invoices
WHERE business_id = ?
  AND created_at >= DATE_SUB(UTC_TIMESTAMP(), INTERVAL 12 MONTH)
GROUP BY status;

-- name: ListOverdueInvoices :many
SELECT id, owner_id, business_id, invoice_number, client_name, client_email, client_address, client_tax_id, customer_id,
       status, issue_date, due_date, paid_date, reporting_date, currency, subtotal, discount_type,
       discount_value, discount_amount, tax_rate, tax_amount, total, amount_paid, amount_due,
       notes, terms, created_at, updated_at
FROM finance_invoices
WHERE business_id = ?
  AND status NOT IN ('paid', 'void', 'draft')
  AND due_date < ?
ORDER BY due_date ASC, id ASC
LIMIT ? OFFSET ?;

-- name: CountOverdueInvoices :one
SELECT COUNT(*)
FROM finance_invoices
WHERE business_id = ?
  AND status NOT IN ('paid', 'void', 'draft')
  AND due_date < ?;

-- name: GetInvoiceItems :many
SELECT id, invoice_id, description, quantity, unit, unit_price, tax_rate, discount_pct, total, sort_order
FROM finance_invoice_items
WHERE invoice_id = ?
ORDER BY sort_order ASC, id ASC
LIMIT 500;

-- name: CreateInvoiceItem :exec
INSERT INTO finance_invoice_items
  (invoice_id, description, quantity, unit, unit_price, tax_rate, discount_pct, total, sort_order)
VALUES (?,?,?,?,?,?,?,?,?);

-- name: DeleteInvoiceItems :exec
DELETE FROM finance_invoice_items WHERE invoice_id = ?;

-- name: CreatePayment :execresult
INSERT INTO finance_payments
  (owner_id, business_id, invoice_id, amount, currency, payment_date, method, reference, notes)
VALUES (?,?,?,?,?,?,?,?,?);

-- name: GetPaymentsByInvoice :many
SELECT id, owner_id, business_id, invoice_id, amount, currency, payment_date, method, reference, notes, created_at
FROM finance_payments
WHERE invoice_id = ?
ORDER BY payment_date ASC, id ASC
LIMIT 500;

-- name: GetPaymentByReference :one
SELECT id, owner_id, business_id, invoice_id, amount, currency, payment_date, method, reference, notes, created_at
FROM finance_payments
WHERE business_id = ? AND invoice_id = ? AND reference = ?;

-- name: ListExpensesByOwner :many
SELECT id, owner_id, business_id, category_id, description, amount, currency, expense_date, vendor,
       payment_method, receipt_url, is_tax_deductible, tax_amount, notes, status, created_at, updated_at
FROM finance_expenses
WHERE business_id = ?
ORDER BY expense_date DESC
LIMIT ? OFFSET ?;

-- name: CountExpensesByOwner :one
SELECT COUNT(*) FROM finance_expenses WHERE business_id = ?;

-- name: GetExpenseByID :one
SELECT id, owner_id, business_id, category_id, description, amount, currency, expense_date, vendor,
       payment_method, receipt_url, is_tax_deductible, tax_amount, notes, status, created_at, updated_at
FROM finance_expenses
WHERE id = ? AND business_id = ?;

-- name: CreateExpense :execresult
INSERT INTO finance_expenses
  (owner_id, business_id, category_id, description, amount, currency, expense_date, vendor,
   payment_method, receipt_url, is_tax_deductible, tax_amount, notes)
VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?);

-- name: UpdateExpense :exec
UPDATE finance_expenses SET
  category_id = ?, description = ?, amount = ?, currency = ?, expense_date = ?,
  vendor = ?, payment_method = ?, receipt_url = ?, is_tax_deductible = ?,
  tax_amount = ?, notes = ?
WHERE id = ? AND business_id = ?;

-- name: DeleteExpense :exec
DELETE FROM finance_expenses WHERE id = ? AND business_id = ?;

-- name: ListExpenseCategories :many
SELECT id, owner_id, business_id, name, color, icon, created_at
FROM finance_expense_categories
WHERE business_id = ?
ORDER BY name ASC
LIMIT 200;

-- name: CreateExpenseCategory :execresult
INSERT INTO finance_expense_categories (owner_id, business_id, name, color, icon)
VALUES (?,?,?,?,?);

-- name: ListTaxConfigs :many
SELECT id, owner_id, business_id, name, rate, type, is_default, is_active, created_at
FROM finance_tax_config
WHERE business_id = ?
ORDER BY is_default DESC, name ASC
LIMIT 200;

-- name: CreateTaxConfig :execresult
INSERT INTO finance_tax_config (owner_id, business_id, name, rate, type, is_default, is_active)
VALUES (?,?,?,?,?,?,?);

-- name: SumPaidInvoices :one
SELECT COALESCE(SUM(total),0) FROM finance_invoices WHERE business_id = ? AND status = 'paid' AND paid_date BETWEEN ? AND ?;

-- name: SumExpenses :one
SELECT COALESCE(SUM(amount),0) FROM finance_expenses WHERE business_id = ? AND expense_date BETWEEN ? AND ?;

-- name: SumTotalReceivable :one
SELECT COALESCE(SUM(amount_due),0) FROM finance_invoices WHERE business_id = ? AND status IN ('sent','viewed','partial','overdue');

-- name: SumOverdueAmount :one
SELECT COALESCE(SUM(amount_due),0) FROM finance_invoices WHERE business_id = ? AND status = 'overdue';

-- name: SumTaxCollected :one
SELECT COALESCE(SUM(tax_amount),0) FROM finance_invoices WHERE business_id = ? AND status = 'paid' AND paid_date BETWEEN ? AND ?;

-- name: SumTaxOwed :one
SELECT COALESCE(SUM(tax_amount),0) FROM finance_expenses WHERE business_id = ? AND is_tax_deductible = TRUE AND expense_date BETWEEN ? AND ?;

-- name: ListInvoicesByCustomer :many
-- Customer-filtered page backing crm.sales-style per-customer views; the
-- index idx_finance_invoices_business_customer serves it.
SELECT id, owner_id, business_id, invoice_number, client_name, client_email, client_address, client_tax_id, customer_id,
       status, issue_date, due_date, paid_date, reporting_date, currency, subtotal, discount_type,
       discount_value, discount_amount, tax_rate, tax_amount, total, amount_paid, amount_due,
       notes, terms, created_at, updated_at
FROM finance_invoices
WHERE business_id = ? AND customer_id = ?
ORDER BY created_at DESC, id DESC
LIMIT ? OFFSET ?;

-- name: CountInvoicesByCustomer :one
SELECT COUNT(*) FROM finance_invoices WHERE business_id = ? AND customer_id = ?;
