-- cross-schema read from accounts — balance-sheet account totals.
-- name: SumAccountsByType :one
SELECT COALESCE(SUM(balance), 0) FROM finance_accounts WHERE business_id = ? AND type = ? AND is_active = TRUE;

-- cross-schema read from bank_accounts — current bank-account balances.
-- name: SumBankBalances :one
SELECT COALESCE(SUM(balance), 0) FROM finance_bank_accounts WHERE business_id = ?;

-- cross-schema read from invoices — outstanding receivables.
-- name: SumReceivables :one
SELECT COALESCE(SUM(amount_due), 0) FROM finance_invoices WHERE business_id = ? AND status NOT IN ('paid','cancelled','void');

-- cross-schema read from bills — outstanding payables.
-- name: SumPayables :one
SELECT COALESCE(SUM(amount), 0) FROM finance_bills WHERE business_id = ? AND status IN ('pending', 'overdue');

-- cross-schema read from invoices — paid revenue for the requested period.
-- name: SumPaidInvoicesInRange :one
SELECT COALESCE(SUM(total), 0) FROM finance_invoices
WHERE business_id = ? AND status = 'paid' AND paid_date BETWEEN ? AND ?;

-- cross-schema read from expenses + invoices — expense totals for the requested period.
-- name: SumExpensesInRange :one
SELECT COALESCE(SUM(amount), 0) FROM finance_expenses
WHERE business_id = ? AND status IN ('pending', 'approved') AND expense_date BETWEEN ? AND ?;
