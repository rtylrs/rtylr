//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `InsightsExpensesByCategory` (Go `InsightsExpensesByCategoryParams`).
pub struct InsightsExpensesByCategoryParams {
    pub business_id: i64,
    pub date_from: chrono::NaiveDate,
    pub date_to: chrono::NaiveDate,
}

/// `InsightsExpensesByCategory`: `many`.
pub async fn insights_expenses_by_category<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsightsExpensesByCategoryParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT COALESCE(c.id, 0) AS cat_id,\n       COALESCE(c.name, 'Uncategorised') AS cat_name,\n       COALESCE(SUM(e.amount), 0) AS amount\nFROM finance_expenses e\nLEFT JOIN finance_expense_categories c ON c.id = e.category_id AND c.business_id = e.business_id\nWHERE e.business_id = ?\n  AND e.status = 'approved'\n  AND e.expense_date BETWEEN ? AND ?\nGROUP BY c.id, c.name\nORDER BY SUM(e.amount) DESC")
        .bind(params.business_id)
        .bind(params.date_from)
        .bind(params.date_to)
        .execute(exec)
        .await
}

/// Params for `InsightsExpensesByDay` (Go `InsightsExpensesByDayParams`).
pub struct InsightsExpensesByDayParams {
    pub business_id: i64,
    pub date_from: chrono::NaiveDate,
    pub date_to: chrono::NaiveDate,
}

/// `InsightsExpensesByDay`: `many`.
pub async fn insights_expenses_by_day<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsightsExpensesByDayParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT DATE(expense_date) AS day,\n       COALESCE(SUM(amount), 0) AS amount\nFROM finance_expenses\nWHERE business_id = ?\n  AND status = 'approved'\n  AND expense_date BETWEEN ? AND ?\nGROUP BY day")
        .bind(params.business_id)
        .bind(params.date_from)
        .bind(params.date_to)
        .execute(exec)
        .await
}

/// Params for `InsightsFinanceLastChange` (Go `InsightsFinanceLastChangeParams`).
pub struct InsightsFinanceLastChangeParams {
    pub business_id: i64,
}

/// `InsightsFinanceLastChange`: `one`.
pub async fn insights_finance_last_change<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsightsFinanceLastChangeParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT CAST(GREATEST(\n  COALESCE((SELECT MAX(i.updated_at) FROM finance_invoices i WHERE i.business_id = ?), '1000-01-01'),\n  COALESCE((SELECT MAX(e.updated_at) FROM finance_expenses e WHERE e.business_id = ?), '1000-01-01')\n) AS DATETIME) AS last_change")
        .bind(params.business_id)
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `InsightsInvoiceAging` (Go `InsightsInvoiceAgingParams`).
pub struct InsightsInvoiceAgingParams {
    pub business_id: i64,
}

/// `InsightsInvoiceAging`: `many`.
pub async fn insights_invoice_aging<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsightsInvoiceAgingParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT\n  CASE\n    WHEN DATEDIFF(CURDATE(), due_date) <= 0 THEN 'current'\n    WHEN DATEDIFF(CURDATE(), due_date) BETWEEN 1 AND 30 THEN '1-30'\n    WHEN DATEDIFF(CURDATE(), due_date) BETWEEN 31 AND 60 THEN '31-60'\n    WHEN DATEDIFF(CURDATE(), due_date) BETWEEN 61 AND 90 THEN '61-90'\n    WHEN DATEDIFF(CURDATE(), due_date) > 90 THEN '90+'\n    ELSE 'current'\n  END AS bucket,\n  COUNT(*) AS invoice_count,\n  COALESCE(SUM(amount_due), 0) AS amount_due\nFROM finance_invoices\nWHERE business_id = ?\n  AND created_at >= DATE_SUB(UTC_TIMESTAMP(), INTERVAL 12 MONTH)\n  AND status NOT IN ('paid','cancelled','void')\n  AND amount_due > 0\nGROUP BY bucket")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `InsightsPaidInvoicesByDay` (Go `InsightsPaidInvoicesByDayParams`).
pub struct InsightsPaidInvoicesByDayParams {
    pub business_id: i64,
    pub date_from: Option<chrono::DateTime<chrono::Utc>>,
    pub date_to: Option<chrono::DateTime<chrono::Utc>>,
}

/// `InsightsPaidInvoicesByDay`: `many`.
pub async fn insights_paid_invoices_by_day<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsightsPaidInvoicesByDayParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT reporting_date AS day,\n       COALESCE(SUM(total), 0) AS amount\nFROM finance_invoices\nWHERE business_id = ?\n  AND status = 'paid'\n  AND reporting_date >= ?\n  AND reporting_date <= ?\nGROUP BY day\nORDER BY day ASC\nLIMIT 200")
        .bind(params.business_id)
        .bind(params.date_from)
        .bind(params.date_to)
        .execute(exec)
        .await
}

/// Params for `ListBalanceSheetAccounts` (Go `ListBalanceSheetAccountsParams`).
pub struct ListBalanceSheetAccountsParams {
    pub business_id: i64,
    pub as_of: chrono::NaiveDate,
    pub reporting_currency: Option<String>,
}

/// `ListBalanceSheetAccounts`: `many`.
pub async fn list_balance_sheet_accounts<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListBalanceSheetAccountsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT a.id, COALESCE(a.code, '') AS code, a.name, a.type,\n       CAST(COALESCE(a.balance, 0) + CASE\n           WHEN a.type = 'asset' THEN COALESCE(p.debit, 0) - COALESCE(p.credit, 0)\n           ELSE COALESCE(p.credit, 0) - COALESCE(p.debit, 0)\n       END AS DECIMAL(15,2)) AS amount\nFROM finance_accounts a\nLEFT JOIN (\n    SELECT l.account_id,\n           SUM(CAST(l.debit AS DECIMAL(15,2))) AS debit,\n           SUM(CAST(l.credit AS DECIMAL(15,2))) AS credit\n    FROM finance_journal_lines l\n    JOIN finance_journal_entries e ON e.id = l.entry_id\n    WHERE e.business_id = ?\n      AND e.status = 'posted'\n      AND e.date <= ?\n    GROUP BY l.account_id\n) p ON p.account_id = a.id\nWHERE a.business_id = ?\n  AND a.type IN ('asset', 'liability', 'equity')\n  AND COALESCE(NULLIF(a.currency, ''), 'USD') = ?\n  AND a.is_active = TRUE\nORDER BY a.type ASC, a.code ASC, a.name ASC")
        .bind(params.business_id)
        .bind(params.as_of)
        .bind(params.business_id)
        .bind(&params.reporting_currency)
        .execute(exec)
        .await
}

/// Params for `SumAccountsByType` (Go `SumAccountsByTypeParams`).
pub struct SumAccountsByTypeParams {
    pub business_id: i64,
    pub account_type: String,
    pub reporting_currency: Option<String>,
}

/// `SumAccountsByType`: `one`.
pub async fn sum_accounts_by_type<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &SumAccountsByTypeParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COALESCE(SUM(\n    COALESCE(a.balance, 0) + CASE\n        WHEN a.type IN ('asset', 'expense') THEN COALESCE(p.debit, 0) - COALESCE(p.credit, 0)\n        ELSE COALESCE(p.credit, 0) - COALESCE(p.debit, 0)\n    END\n), 0)\nFROM finance_accounts a\nLEFT JOIN (\n    SELECT l.account_id,\n           SUM(CAST(l.debit AS DECIMAL(15,2))) AS debit,\n           SUM(CAST(l.credit AS DECIMAL(15,2))) AS credit\n    FROM finance_journal_lines l\n    JOIN finance_journal_entries e ON e.id = l.entry_id\n    WHERE e.business_id = ?\n      AND e.status = 'posted'\n    GROUP BY l.account_id\n) p ON p.account_id = a.id\nWHERE a.business_id = ?\n  AND a.type = ?\n  AND COALESCE(NULLIF(a.currency, ''), 'USD') = ?\n  AND a.is_active = TRUE")
        .bind(params.business_id)
        .bind(params.business_id)
        .bind(&params.account_type)
        .bind(&params.reporting_currency)
        .fetch_optional(exec)
        .await
}

/// Params for `SumBankBalances` (Go `SumBankBalancesParams`).
pub struct SumBankBalancesParams {
    pub business_id: i64,
    pub reporting_currency: Option<String>,
}

/// `SumBankBalances`: `one`.
pub async fn sum_bank_balances<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &SumBankBalancesParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COALESCE(SUM(a.balance), 0)\nFROM finance_bank_accounts a\nWHERE a.business_id = ?\n  AND COALESCE(NULLIF(a.currency, ''), 'USD') = ?")
        .bind(params.business_id)
        .bind(&params.reporting_currency)
        .fetch_optional(exec)
        .await
}

/// Params for `SumExpensesInRange` (Go `SumExpensesInRangeParams`).
pub struct SumExpensesInRangeParams {
    pub business_id: i64,
    pub from_expense_date: chrono::NaiveDate,
    pub to_expense_date: chrono::NaiveDate,
    pub reporting_currency: Option<String>,
}

/// `SumExpensesInRange`: `one`.
pub async fn sum_expenses_in_range<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &SumExpensesInRangeParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COALESCE(SUM(e.amount), 0)\nFROM finance_expenses e\nWHERE e.business_id = ?\n  AND e.status = 'approved'\n  AND e.expense_date BETWEEN ? AND ?\n  AND COALESCE(NULLIF(e.currency, ''), 'USD') = ?")
        .bind(params.business_id)
        .bind(params.from_expense_date)
        .bind(params.to_expense_date)
        .bind(&params.reporting_currency)
        .fetch_optional(exec)
        .await
}

/// Params for `SumInsightsExpenses` (Go `SumInsightsExpensesParams`).
pub struct SumInsightsExpensesParams {
    pub business_id: i64,
    pub date_from: chrono::NaiveDate,
    pub date_to: chrono::NaiveDate,
}

/// `SumInsightsExpenses`: `one`.
pub async fn sum_insights_expenses<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &SumInsightsExpensesParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COALESCE(SUM(amount), 0)\nFROM finance_expenses\nWHERE business_id = ?\n  AND status = 'approved'\n  AND expense_date >= ?\n  AND expense_date <= ?")
        .bind(params.business_id)
        .bind(params.date_from)
        .bind(params.date_to)
        .fetch_optional(exec)
        .await
}

/// Params for `SumInsightsInvoiceRevenue` (Go `SumInsightsInvoiceRevenueParams`).
pub struct SumInsightsInvoiceRevenueParams {
    pub business_id: i64,
    pub date_from: Option<chrono::DateTime<chrono::Utc>>,
    pub date_to: Option<chrono::DateTime<chrono::Utc>>,
}

/// `SumInsightsInvoiceRevenue`: `one`.
pub async fn sum_insights_invoice_revenue<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &SumInsightsInvoiceRevenueParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COALESCE(SUM(total), 0) AS total\nFROM finance_invoices\nWHERE business_id = ?\n  AND status = 'paid'\n  AND reporting_date >= ?\n  AND reporting_date <= ?")
        .bind(params.business_id)
        .bind(params.date_from)
        .bind(params.date_to)
        .fetch_optional(exec)
        .await
}

/// Params for `SumInsightsPaidInvoices` (Go `SumInsightsPaidInvoicesParams`).
pub struct SumInsightsPaidInvoicesParams {
    pub business_id: i64,
    pub date_from: Option<chrono::DateTime<chrono::Utc>>,
    pub date_to: Option<chrono::DateTime<chrono::Utc>>,
}

/// `SumInsightsPaidInvoices`: `one`.
pub async fn sum_insights_paid_invoices<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &SumInsightsPaidInvoicesParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COALESCE(SUM(total), 0)\nFROM finance_invoices\nWHERE business_id = ?\n  AND status = 'paid'\n  AND COALESCE(paid_date, issue_date) >= ?\n  AND COALESCE(paid_date, issue_date) <= ?")
        .bind(params.business_id)
        .bind(params.date_from)
        .bind(params.date_to)
        .fetch_optional(exec)
        .await
}

/// Params for `SumPaidInvoicesInRange` (Go `SumPaidInvoicesInRangeParams`).
pub struct SumPaidInvoicesInRangeParams {
    pub business_id: i64,
    pub from_paid_date: Option<chrono::DateTime<chrono::Utc>>,
    pub to_paid_date: Option<chrono::DateTime<chrono::Utc>>,
    pub reporting_currency: Option<String>,
}

/// `SumPaidInvoicesInRange`: `one`.
pub async fn sum_paid_invoices_in_range<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &SumPaidInvoicesInRangeParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COALESCE(SUM(i.total), 0)\nFROM finance_invoices i\nWHERE i.business_id = ?\n  AND i.status = 'paid'\n  AND i.paid_date BETWEEN ? AND ?\n  AND COALESCE(NULLIF(i.currency, ''), 'USD') = ?")
        .bind(params.business_id)
        .bind(params.from_paid_date)
        .bind(params.to_paid_date)
        .bind(&params.reporting_currency)
        .fetch_optional(exec)
        .await
}

/// Params for `SumPayables` (Go `SumPayablesParams`).
pub struct SumPayablesParams {
    pub business_id: i64,
    pub reporting_currency: Option<String>,
}

/// `SumPayables`: `one`.
pub async fn sum_payables<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &SumPayablesParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COALESCE(SUM(bill.amount), 0)\nFROM finance_bills bill\nWHERE bill.business_id = ?\n  AND bill.status IN ('pending', 'overdue')\n  AND COALESCE(NULLIF(bill.currency, ''), 'USD') = ?")
        .bind(params.business_id)
        .bind(&params.reporting_currency)
        .fetch_optional(exec)
        .await
}

/// Params for `SumPostedNetIncome` (Go `SumPostedNetIncomeParams`).
pub struct SumPostedNetIncomeParams {
    pub business_id: i64,
    pub reporting_currency: Option<String>,
}

/// `SumPostedNetIncome`: `one`.
pub async fn sum_posted_net_income<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &SumPostedNetIncomeParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COALESCE(SUM(\n    CASE\n        WHEN a.type = 'revenue' THEN CAST(l.credit AS DECIMAL(15,2)) - CAST(l.debit AS DECIMAL(15,2))\n        WHEN a.type = 'expense' THEN CAST(l.credit AS DECIMAL(15,2)) - CAST(l.debit AS DECIMAL(15,2))\n        ELSE 0\n    END\n), 0)\nFROM finance_journal_lines l\nJOIN finance_journal_entries e ON e.id = l.entry_id AND e.status = 'posted'\nJOIN finance_accounts a ON a.id = l.account_id AND a.business_id = e.business_id\nWHERE e.business_id = ?\n  AND a.type IN ('revenue', 'expense')\n  AND COALESCE(NULLIF(a.currency, ''), 'USD') = ?\n  AND a.is_active = TRUE")
        .bind(params.business_id)
        .bind(&params.reporting_currency)
        .fetch_optional(exec)
        .await
}

/// Params for `SumPostedNetIncomeAsOf` (Go `SumPostedNetIncomeAsOfParams`).
pub struct SumPostedNetIncomeAsOfParams {
    pub business_id: i64,
    pub as_of: chrono::NaiveDate,
    pub reporting_currency: Option<String>,
}

/// `SumPostedNetIncomeAsOf`: `one`.
pub async fn sum_posted_net_income_as_of<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &SumPostedNetIncomeAsOfParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COALESCE(SUM(CAST(l.credit AS DECIMAL(15,2)) - CAST(l.debit AS DECIMAL(15,2))), 0)\nFROM finance_journal_lines l\nJOIN finance_journal_entries e ON e.id = l.entry_id AND e.status = 'posted'\nJOIN finance_accounts a ON a.id = l.account_id AND a.business_id = e.business_id\nWHERE e.business_id = ?\n  AND e.date <= ?\n  AND a.type IN ('revenue', 'expense')\n  AND COALESCE(NULLIF(a.currency, ''), 'USD') = ?\n  AND a.is_active = TRUE")
        .bind(params.business_id)
        .bind(params.as_of)
        .bind(&params.reporting_currency)
        .fetch_optional(exec)
        .await
}

/// Params for `SumReceivables` (Go `SumReceivablesParams`).
pub struct SumReceivablesParams {
    pub business_id: i64,
    pub reporting_currency: Option<String>,
}

/// `SumReceivables`: `one`.
pub async fn sum_receivables<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &SumReceivablesParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COALESCE(SUM(i.amount_due), 0)\nFROM finance_invoices i\nWHERE i.business_id = ?\n  AND i.status NOT IN ('paid','cancelled','void')\n  AND COALESCE(NULLIF(i.currency, ''), 'USD') = ?")
        .bind(params.business_id)
        .bind(&params.reporting_currency)
        .fetch_optional(exec)
        .await
}
