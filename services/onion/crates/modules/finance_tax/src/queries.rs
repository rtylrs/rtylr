//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `ClearDefaultTaxConfig` (Go `ClearDefaultTaxConfigParams`).
pub struct ClearDefaultTaxConfigParams {
    pub business_id: i64,
}

/// `ClearDefaultTaxConfig`: `exec`.
pub async fn clear_default_tax_config<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ClearDefaultTaxConfigParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE finance_tax_config SET is_default = FALSE WHERE business_id = ?")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `GetTaxConfig` (Go `GetTaxConfigParams`).
pub struct GetTaxConfigParams {
    pub business_id: i64,
}

/// `GetTaxConfig`: `one`.
pub async fn get_tax_config<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetTaxConfigParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, name, rate, type, is_default, is_active, is_inclusive, created_at\nFROM finance_tax_config\nWHERE business_id = ? AND is_default = TRUE\nORDER BY id ASC\nLIMIT 1")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `SumAllExpensesAmount` (Go `SumAllExpensesAmountParams`).
pub struct SumAllExpensesAmountParams {
    pub business_id: i64,
}

/// `SumAllExpensesAmount`: `one`.
pub async fn sum_all_expenses_amount<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &SumAllExpensesAmountParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COALESCE(SUM(amount), 0) FROM finance_expenses WHERE business_id = ? AND status = 'approved'")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `SumPaidInvoicesTotal` (Go `SumPaidInvoicesTotalParams`).
pub struct SumPaidInvoicesTotalParams {
    pub business_id: i64,
}

/// `SumPaidInvoicesTotal`: `one`.
pub async fn sum_paid_invoices_total<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &SumPaidInvoicesTotalParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COALESCE(SUM(total), 0) FROM finance_invoices WHERE business_id = ? AND status = 'paid'")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `UpsertTaxConfig` (Go `UpsertTaxConfigParams`).
pub struct UpsertTaxConfigParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub name: String,
    pub rate: String,
    pub r#type: String,
    pub is_active: Option<bool>,
    pub is_inclusive: Option<bool>,
}

/// `UpsertTaxConfig`: `exec`.
pub async fn upsert_tax_config<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpsertTaxConfigParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO finance_tax_config (owner_id, business_id, name, rate, type, is_default, is_active, is_inclusive)\nVALUES (?, ?, ?, ?, ?, TRUE, ?, ?)\nON DUPLICATE KEY UPDATE\n  name = VALUES(name),\n  rate = VALUES(rate),\n  is_default = TRUE,\n  is_active = VALUES(is_active),\n  is_inclusive = VALUES(is_inclusive)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(&params.name)
        .bind(&params.rate)
        .bind(&params.r#type)
        .bind(params.is_active)
        .bind(params.is_inclusive)
        .execute(exec)
        .await
}
