//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `AdvanceInvoiceSequence` (Go `AdvanceInvoiceSequenceParams`).
pub struct AdvanceInvoiceSequenceParams {
    pub last_sequence: i64,
    pub business_id: i64,
    pub prefix: String,
}

/// `AdvanceInvoiceSequence`: `exec`.
pub async fn advance_invoice_sequence<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &AdvanceInvoiceSequenceParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE finance_invoice_sequences SET last_sequence = ?\nWHERE business_id = ? AND prefix = ?")
        .bind(params.last_sequence)
        .bind(params.business_id)
        .bind(&params.prefix)
        .execute(exec)
        .await
}

/// Params for `EnsureInvoiceSequence` (Go `EnsureInvoiceSequenceParams`).
pub struct EnsureInvoiceSequenceParams {
    pub business_id: i64,
    pub prefix: String,
}

/// `EnsureInvoiceSequence`: `exec`.
pub async fn ensure_invoice_sequence<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &EnsureInvoiceSequenceParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO finance_invoice_sequences (business_id, prefix, last_sequence)\nVALUES (?, ?, 0)\nON DUPLICATE KEY UPDATE last_sequence = last_sequence")
        .bind(params.business_id)
        .bind(&params.prefix)
        .execute(exec)
        .await
}

/// Params for `LockInvoiceSequence` (Go `LockInvoiceSequenceParams`).
pub struct LockInvoiceSequenceParams {
    pub business_id: i64,
    pub prefix: String,
}

/// `LockInvoiceSequence`: `one`.
pub async fn lock_invoice_sequence<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &LockInvoiceSequenceParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT last_sequence FROM finance_invoice_sequences\nWHERE business_id = ? AND prefix = ? FOR UPDATE")
        .bind(params.business_id)
        .bind(&params.prefix)
        .fetch_optional(exec)
        .await
}

/// Params for `MaxInvoiceSequence` (Go `MaxInvoiceSequenceParams`).
pub struct MaxInvoiceSequenceParams {
    pub prefix: String,
    pub business_id: i64,
}

/// `MaxInvoiceSequence`: `one`.
pub async fn max_invoice_sequence<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &MaxInvoiceSequenceParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT CAST(COALESCE(MAX(CAST(SUBSTRING(invoice_number, LENGTH(?) + 1) AS UNSIGNED)), 0) AS SIGNED) AS sequence\nFROM finance_invoices\nWHERE business_id = ? AND invoice_number LIKE CONCAT(?, '%')\n  AND SUBSTRING_INDEX(invoice_number, '-', -1) REGEXP '^[0-9]+$'")
        .bind(&params.prefix)
        .bind(params.business_id)
        .bind(&params.prefix)
        .fetch_optional(exec)
        .await
}
