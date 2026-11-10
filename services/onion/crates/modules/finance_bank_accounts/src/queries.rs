//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CountConnectedBankAccounts` (Go `CountConnectedBankAccountsParams`).
pub struct CountConnectedBankAccountsParams {
    pub business_id: i64,
}

/// `CountConnectedBankAccounts`: `one`.
pub async fn count_connected_bank_accounts<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountConnectedBankAccountsParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*)\nFROM finance_bank_accounts\nWHERE business_id = ? AND is_connected = TRUE")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}
