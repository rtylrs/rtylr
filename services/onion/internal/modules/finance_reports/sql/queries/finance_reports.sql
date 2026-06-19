-- name: SumAccountsByType :one
SELECT COALESCE(SUM(balance), 0) FROM finance_accounts WHERE business_id = ? AND type = ? AND is_active = TRUE;

-- name: SumBankBalances :one
SELECT COALESCE(SUM(balance), 0) FROM finance_bank_accounts WHERE business_id = ?;

-- name: SumReceivables :one
SELECT COALESCE(SUM(amount_due), 0) FROM finance_invoices WHERE business_id = ? AND status NOT IN ('paid','cancelled','void');

-- name: SumPayables :one
SELECT COALESCE(SUM(amount), 0) FROM finance_bills WHERE business_id = ? AND status = 'pending';

-- name: SumPaidInvoicesInRange :one
SELECT COALESCE(SUM(total), 0) FROM finance_invoices
WHERE business_id = ? AND status = 'paid' AND paid_date BETWEEN ? AND ?;

-- name: SumExpensesInRange :one
SELECT COALESCE(SUM(amount), 0) FROM finance_expenses
WHERE business_id = ? AND expense_date BETWEEN ? AND ?;
