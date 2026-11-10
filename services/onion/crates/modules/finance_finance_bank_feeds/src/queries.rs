//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CountBankFeedTransactions` (Go `CountBankFeedTransactionsParams`).
pub struct CountBankFeedTransactionsParams {
    pub business_id: i64,
}

/// `CountBankFeedTransactions`: `one`.
pub async fn count_bank_feed_transactions<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountBankFeedTransactionsParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM finance_bank_feeds WHERE business_id = ?")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CountUnmatchedBankFeedTransactions` (Go `CountUnmatchedBankFeedTransactionsParams`).
pub struct CountUnmatchedBankFeedTransactionsParams {
    pub business_id: i64,
}

/// `CountUnmatchedBankFeedTransactions`: `one`.
pub async fn count_unmatched_bank_feed_transactions<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountUnmatchedBankFeedTransactionsParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*)\nFROM finance_bank_feeds\nWHERE business_id = ?\n  AND matched_expense_id IS NULL\n  AND matched_invoice_id IS NULL")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `ListSyncBankAccounts` (Go `ListSyncBankAccountsParams`).
pub struct ListSyncBankAccountsParams {
    pub business_id: i64,
}

/// `ListSyncBankAccounts`: `many`.
pub async fn list_sync_bank_accounts<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListSyncBankAccountsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, balance, last_synced_at\nFROM finance_bank_accounts\nWHERE business_id = ? AND is_connected = TRUE\nFOR UPDATE")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `SumBankFeedAmountByAccount` (Go `SumBankFeedAmountByAccountParams`).
pub struct SumBankFeedAmountByAccountParams {
    pub business_id: i64,
    pub bank_account_id: i64,
}

/// `SumBankFeedAmountByAccount`: `one`.
pub async fn sum_bank_feed_amount_by_account<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &SumBankFeedAmountByAccountParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COALESCE(SUM(amount_minor_units), 0)\nFROM finance_bank_feeds\nWHERE business_id = ? AND bank_account_id = ?")
        .bind(params.business_id)
        .bind(params.bank_account_id)
        .fetch_optional(exec)
        .await
}

/// Params for `UpdateSyncedBankAccount` (Go `UpdateSyncedBankAccountParams`).
pub struct UpdateSyncedBankAccountParams {
    pub balance: Option<String>,
    pub id: i64,
    pub business_id: i64,
}

/// `UpdateSyncedBankAccount`: `execresult`.
pub async fn update_synced_bank_account<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateSyncedBankAccountParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE finance_bank_accounts\nSET balance = ?, last_synced_at = CURRENT_TIMESTAMP\nWHERE id = ? AND business_id = ?")
        .bind(&params.balance)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}
