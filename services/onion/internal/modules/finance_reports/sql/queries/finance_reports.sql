-- cross-schema read from accounts — balance-sheet account totals.
-- name: SumAccountsByType :one
SELECT COALESCE(SUM(
    COALESCE(a.balance, 0) + CASE
        WHEN a.type IN ('asset', 'expense') THEN COALESCE(p.debit, 0) - COALESCE(p.credit, 0)
        ELSE COALESCE(p.credit, 0) - COALESCE(p.debit, 0)
    END
), 0)
FROM finance_accounts a
JOIN businesses b ON b.id = a.business_id
LEFT JOIN (
    SELECT l.account_id,
           SUM(CAST(l.debit AS DECIMAL(15,2))) AS debit,
           SUM(CAST(l.credit AS DECIMAL(15,2))) AS credit
    FROM finance_journal_lines l
    JOIN finance_journal_entries e ON e.id = l.entry_id
    WHERE e.business_id = sqlc.arg(business_id)
      AND e.status = 'posted'
    GROUP BY l.account_id
) p ON p.account_id = a.id
WHERE a.business_id = sqlc.arg(business_id)
  AND a.type = sqlc.arg(account_type)
  AND COALESCE(NULLIF(a.currency, ''), 'USD') = COALESCE(NULLIF(b.currency, ''), 'USD')
  AND a.is_active = TRUE;

-- cross-schema read from journal — cumulative posted revenue and expense
-- balances become the contribution to retained earnings.
-- name: SumPostedNetIncome :one
SELECT COALESCE(SUM(
    CASE
        WHEN a.type = 'revenue' THEN CAST(l.credit AS DECIMAL(15,2)) - CAST(l.debit AS DECIMAL(15,2))
        WHEN a.type = 'expense' THEN CAST(l.debit AS DECIMAL(15,2)) - CAST(l.credit AS DECIMAL(15,2))
        ELSE 0
    END
), 0)
FROM finance_journal_lines l
JOIN finance_journal_entries e ON e.id = l.entry_id AND e.status = 'posted'
JOIN finance_accounts a ON a.id = l.account_id AND a.business_id = e.business_id
JOIN businesses b ON b.id = e.business_id
WHERE e.business_id = sqlc.arg(business_id)
  AND a.type IN ('revenue', 'expense')
  AND COALESCE(NULLIF(a.currency, ''), 'USD') = COALESCE(NULLIF(b.currency, ''), 'USD')
  AND a.is_active = TRUE;

-- cross-schema read from bank_accounts — current bank-account balances.
-- name: SumBankBalances :one
SELECT COALESCE(SUM(a.balance), 0)
FROM finance_bank_accounts a
JOIN businesses b ON b.id = a.business_id
WHERE a.business_id = ?
  AND COALESCE(NULLIF(a.currency, ''), 'USD') = COALESCE(NULLIF(b.currency, ''), 'USD');

-- cross-schema read from invoices — outstanding receivables.
-- name: SumReceivables :one
SELECT COALESCE(SUM(i.amount_due), 0)
FROM finance_invoices i
JOIN businesses b ON b.id = i.business_id
WHERE i.business_id = ?
  AND i.status NOT IN ('paid','cancelled','void')
  AND COALESCE(NULLIF(i.currency, ''), 'USD') = COALESCE(NULLIF(b.currency, ''), 'USD');

-- cross-schema read from bills — outstanding payables.
-- name: SumPayables :one
SELECT COALESCE(SUM(bill.amount), 0)
FROM finance_bills bill
JOIN businesses b ON b.id = bill.business_id
WHERE bill.business_id = ?
  AND bill.status IN ('pending', 'overdue')
  AND COALESCE(NULLIF(bill.currency, ''), 'USD') = COALESCE(NULLIF(b.currency, ''), 'USD');

-- cross-schema read from invoices — paid revenue for the requested period.
-- name: SumPaidInvoicesInRange :one
SELECT COALESCE(SUM(i.total), 0)
FROM finance_invoices i
JOIN businesses b ON b.id = i.business_id
WHERE i.business_id = ?
  AND i.status = 'paid'
  AND i.paid_date BETWEEN ? AND ?
  AND COALESCE(NULLIF(i.currency, ''), 'USD') = COALESCE(NULLIF(b.currency, ''), 'USD');

-- cross-schema read from expenses + invoices — expense totals for the requested period.
-- name: SumExpensesInRange :one
SELECT COALESCE(SUM(e.amount), 0)
FROM finance_expenses e
JOIN businesses b ON b.id = e.business_id
WHERE e.business_id = ?
  AND e.status IN ('pending', 'approved')
  AND e.expense_date BETWEEN ? AND ?
  AND COALESCE(NULLIF(e.currency, ''), 'USD') = COALESCE(NULLIF(b.currency, ''), 'USD');
