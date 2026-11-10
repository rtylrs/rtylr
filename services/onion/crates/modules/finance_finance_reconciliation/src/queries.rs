//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CountMatches` (Go `CountMatchesParams`).
pub struct CountMatchesParams {
    pub business_id: i64,
}

/// `CountMatches`: `one`.
pub async fn count_matches<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountMatchesParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*)\nFROM finance_reconciliation_matches\nWHERE business_id = ?")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CountMatchesByTargetType` (Go `CountMatchesByTargetTypeParams`).
pub struct CountMatchesByTargetTypeParams {
    pub business_id: i64,
}

/// `CountMatchesByTargetType`: `many`.
pub async fn count_matches_by_target_type<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountMatchesByTargetTypeParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT target_type, COUNT(*) AS cnt\nFROM finance_reconciliation_matches\nWHERE business_id = ?\n  AND created_at >= DATE_SUB(UTC_TIMESTAMP(), INTERVAL 12 MONTH)\nGROUP BY target_type")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `CreateReconciliationMatch` (Go `CreateReconciliationMatchParams`).
pub struct CreateReconciliationMatchParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub bank_feed_id: i64,
    pub target_type: String,
    pub target_id: i64,
    pub matched_by: Option<i64>,
}

/// `CreateReconciliationMatch`: `execresult`.
pub async fn create_reconciliation_match<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateReconciliationMatchParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO finance_reconciliation_matches (\n    owner_id, business_id, bank_feed_id, target_type, target_id, matched_by\n) VALUES (?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.bank_feed_id)
        .bind(&params.target_type)
        .bind(params.target_id)
        .bind(params.matched_by)
        .execute(exec)
        .await
}

/// Params for `GetBankFeedForOwner` (Go `GetBankFeedForOwnerParams`).
pub struct GetBankFeedForOwnerParams {
    pub id: i64,
    pub business_id: i64,
}

/// `GetBankFeedForOwner`: `one`.
pub async fn get_bank_feed_for_owner<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetBankFeedForOwnerParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id FROM finance_bank_feeds\nWHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetExpenseForBusiness` (Go `GetExpenseForBusinessParams`).
pub struct GetExpenseForBusinessParams {
    pub id: i64,
    pub business_id: i64,
}

/// `GetExpenseForBusiness`: `one`.
pub async fn get_expense_for_business<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetExpenseForBusinessParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id FROM finance_expenses\nWHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetInvoiceForBusiness` (Go `GetInvoiceForBusinessParams`).
pub struct GetInvoiceForBusinessParams {
    pub id: i64,
    pub business_id: i64,
}

/// `GetInvoiceForBusiness`: `one`.
pub async fn get_invoice_for_business<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetInvoiceForBusinessParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id FROM finance_invoices\nWHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetMatchForBankFeed` (Go `GetMatchForBankFeedParams`).
pub struct GetMatchForBankFeedParams {
    pub business_id: i64,
    pub bank_feed_id: i64,
}

/// `GetMatchForBankFeed`: `one`.
pub async fn get_match_for_bank_feed<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetMatchForBankFeedParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id FROM finance_reconciliation_matches\nWHERE business_id = ? AND bank_feed_id = ?\nORDER BY id ASC\nLIMIT 1")
        .bind(params.business_id)
        .bind(params.bank_feed_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetMatchForTarget` (Go `GetMatchForTargetParams`).
pub struct GetMatchForTargetParams {
    pub business_id: i64,
    pub target_type: String,
    pub target_id: i64,
}

/// `GetMatchForTarget`: `one`.
pub async fn get_match_for_target<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetMatchForTargetParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id FROM finance_reconciliation_matches\nWHERE business_id = ? AND target_type = ? AND target_id = ?\nORDER BY id ASC\nLIMIT 1")
        .bind(params.business_id)
        .bind(&params.target_type)
        .bind(params.target_id)
        .fetch_optional(exec)
        .await
}

/// Params for `ListUnmatchedBankFeeds` (Go `ListUnmatchedBankFeedsParams`).
pub struct ListUnmatchedBankFeedsParams {
    pub business_id: i64,
}

/// `ListUnmatchedBankFeeds`: `many`.
pub async fn list_unmatched_bank_feeds<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListUnmatchedBankFeedsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT bf.id, bf.owner_id, bf.business_id, bf.bank_account_id, bf.amount_minor_units, bf.currency,\n       bf.occurred_at, bf.description, bf.external_id\nFROM finance_bank_feeds bf\nWHERE bf.business_id = ?\n  AND bf.matched_expense_id IS NULL\n  AND bf.matched_invoice_id IS NULL\n  AND NOT EXISTS (\n      SELECT 1 FROM finance_reconciliation_matches m FORCE INDEX (idx_finance_reconciliation_business_bank)\n      WHERE m.business_id = bf.business_id\n        AND m.bank_feed_id = bf.id\n  )\nORDER BY bf.occurred_at DESC, bf.id DESC\nLIMIT 201")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `ListUnmatchedExpenses` (Go `ListUnmatchedExpensesParams`).
pub struct ListUnmatchedExpensesParams {
    pub business_id: i64,
}

/// `ListUnmatchedExpenses`: `many`.
pub async fn list_unmatched_expenses<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListUnmatchedExpensesParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT e.id, e.description, COALESCE(e.vendor, '') AS vendor, e.amount, e.currency, e.expense_date\nFROM finance_expenses e\nWHERE e.business_id = ?\n  AND e.status <> 'rejected'\n  AND NOT EXISTS (\n      SELECT 1 FROM finance_reconciliation_matches m FORCE INDEX (idx_finance_reconciliation_business_target)\n      WHERE m.business_id = e.business_id\n        AND m.target_type = 'expense'\n        AND m.target_id = e.id\n  )\nORDER BY e.expense_date ASC, e.id ASC\nLIMIT 201")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `ListUnmatchedInvoices` (Go `ListUnmatchedInvoicesParams`).
pub struct ListUnmatchedInvoicesParams {
    pub business_id: i64,
}

/// `ListUnmatchedInvoices`: `many`.
pub async fn list_unmatched_invoices<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListUnmatchedInvoicesParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT i.id, i.invoice_number, i.client_name, i.total, i.currency, i.issue_date, i.due_date\nFROM finance_invoices i\nWHERE i.business_id = ?\n  AND i.status IN ('draft','sent','viewed','partial','overdue')\n  AND NOT EXISTS (\n      SELECT 1 FROM finance_reconciliation_matches m FORCE INDEX (idx_finance_reconciliation_business_target)\n      WHERE m.business_id = i.business_id\n        AND m.target_type = 'invoice'\n        AND m.target_id = i.id\n  )\nORDER BY i.due_date ASC, i.id ASC\nLIMIT 201")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `LockBankFeedForBusiness` (Go `LockBankFeedForBusinessParams`).
pub struct LockBankFeedForBusinessParams {
    pub id: i64,
    pub business_id: i64,
}

/// `LockBankFeedForBusiness`: `one`.
pub async fn lock_bank_feed_for_business<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &LockBankFeedForBusinessParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id FROM finance_bank_feeds\nWHERE id = ? AND business_id = ?\nFOR UPDATE")
        .bind(params.id)
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `LockExpenseForBusiness` (Go `LockExpenseForBusinessParams`).
pub struct LockExpenseForBusinessParams {
    pub id: i64,
    pub business_id: i64,
}

/// `LockExpenseForBusiness`: `one`.
pub async fn lock_expense_for_business<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &LockExpenseForBusinessParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id FROM finance_expenses\nWHERE id = ? AND business_id = ?\nFOR UPDATE")
        .bind(params.id)
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `LockInvoiceForBusiness` (Go `LockInvoiceForBusinessParams`).
pub struct LockInvoiceForBusinessParams {
    pub id: i64,
    pub business_id: i64,
}

/// `LockInvoiceForBusiness`: `one`.
pub async fn lock_invoice_for_business<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &LockInvoiceForBusinessParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id FROM finance_invoices\nWHERE id = ? AND business_id = ?\nFOR UPDATE")
        .bind(params.id)
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `ReserveReconciliationTarget` (Go `ReserveReconciliationTargetParams`).
pub struct ReserveReconciliationTargetParams {
    pub business_id: i64,
    pub target_type: String,
    pub target_id: i64,
}

/// `ReserveReconciliationTarget`: `execresult`.
pub async fn reserve_reconciliation_target<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ReserveReconciliationTargetParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT IGNORE INTO finance_reconciliation_targets (business_id, target_type, target_id)\nVALUES (?, ?, ?)")
        .bind(params.business_id)
        .bind(&params.target_type)
        .bind(params.target_id)
        .execute(exec)
        .await
}

/// Params for `SetBankFeedExpenseMatch` (Go `SetBankFeedExpenseMatchParams`).
pub struct SetBankFeedExpenseMatchParams {
    pub matched_expense_id: Option<i64>,
    pub id: i64,
    pub business_id: i64,
}

/// `SetBankFeedExpenseMatch`: `execresult`.
pub async fn set_bank_feed_expense_match<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &SetBankFeedExpenseMatchParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE finance_bank_feeds\nSET matched_expense_id = ?, matched_invoice_id = NULL\nWHERE id = ? AND business_id = ?\n  AND matched_invoice_id IS NULL\n  AND matched_expense_id IS NULL")
        .bind(params.matched_expense_id)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `SetBankFeedInvoiceMatch` (Go `SetBankFeedInvoiceMatchParams`).
pub struct SetBankFeedInvoiceMatchParams {
    pub matched_invoice_id: Option<i64>,
    pub id: i64,
    pub business_id: i64,
}

/// `SetBankFeedInvoiceMatch`: `execresult`.
pub async fn set_bank_feed_invoice_match<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &SetBankFeedInvoiceMatchParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE finance_bank_feeds\nSET matched_invoice_id = ?, matched_expense_id = NULL\nWHERE id = ? AND business_id = ?\n  AND matched_invoice_id IS NULL\n  AND matched_expense_id IS NULL")
        .bind(params.matched_invoice_id)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}
