-- name: SumAccountsByType :one
SELECT COALESCE(SUM(
    COALESCE(a.balance, 0) + CASE
        WHEN a.type IN ('asset', 'expense') THEN COALESCE(p.debit, 0) - COALESCE(p.credit, 0)
        ELSE COALESCE(p.credit, 0) - COALESCE(p.debit, 0)
    END
), 0)
FROM finance_accounts a
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
  AND COALESCE(NULLIF(a.currency, ''), 'USD') = sqlc.arg(reporting_currency)
  AND a.is_active = TRUE;

-- name: SumPostedNetIncome :one
SELECT COALESCE(SUM(
    CASE
        WHEN a.type = 'revenue' THEN CAST(l.credit AS DECIMAL(15,2)) - CAST(l.debit AS DECIMAL(15,2))
        WHEN a.type = 'expense' THEN CAST(l.credit AS DECIMAL(15,2)) - CAST(l.debit AS DECIMAL(15,2))
        ELSE 0
    END
), 0)
FROM finance_journal_lines l
JOIN finance_journal_entries e ON e.id = l.entry_id AND e.status = 'posted'
JOIN finance_accounts a ON a.id = l.account_id AND a.business_id = e.business_id
WHERE e.business_id = sqlc.arg(business_id)
  AND a.type IN ('revenue', 'expense')
  AND COALESCE(NULLIF(a.currency, ''), 'USD') = sqlc.arg(reporting_currency)
  AND a.is_active = TRUE;

-- name: SumBankBalances :one
SELECT COALESCE(SUM(a.balance), 0)
FROM finance_bank_accounts a
WHERE a.business_id = sqlc.arg(business_id)
  AND COALESCE(NULLIF(a.currency, ''), 'USD') = sqlc.arg(reporting_currency);

-- name: SumReceivables :one
SELECT COALESCE(SUM(i.amount_due), 0)
FROM finance_invoices i
WHERE i.business_id = sqlc.arg(business_id)
  AND i.status NOT IN ('paid','cancelled','void')
  AND COALESCE(NULLIF(i.currency, ''), 'USD') = sqlc.arg(reporting_currency);

-- name: SumPayables :one
SELECT COALESCE(SUM(bill.amount), 0)
FROM finance_bills bill
WHERE bill.business_id = sqlc.arg(business_id)
  AND bill.status IN ('pending', 'overdue')
  AND COALESCE(NULLIF(bill.currency, ''), 'USD') = sqlc.arg(reporting_currency);

-- name: SumPaidInvoicesInRange :one
SELECT COALESCE(SUM(i.total), 0)
FROM finance_invoices i
WHERE i.business_id = sqlc.arg(business_id)
  AND i.status = 'paid'
  AND i.paid_date BETWEEN ? AND ?
  AND COALESCE(NULLIF(i.currency, ''), 'USD') = sqlc.arg(reporting_currency);

-- name: SumExpensesInRange :one
SELECT COALESCE(SUM(e.amount), 0)
FROM finance_expenses e
WHERE e.business_id = sqlc.arg(business_id)
  AND e.status = 'approved'
  AND e.expense_date BETWEEN ? AND ?
  AND COALESCE(NULLIF(e.currency, ''), 'USD') = sqlc.arg(reporting_currency);
