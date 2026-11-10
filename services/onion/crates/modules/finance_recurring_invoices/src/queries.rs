//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CreateRecurringInvoice` (Go `CreateRecurringInvoiceParams`).
pub struct CreateRecurringInvoiceParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub template_invoice_id: i64,
    pub frequency: String,
    pub next_invoice_date: chrono::NaiveDate,
    pub end_date: Option<chrono::DateTime<chrono::Utc>>,
    pub is_active: Option<bool>,
}

/// `CreateRecurringInvoice`: `execresult`.
pub async fn create_recurring_invoice<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateRecurringInvoiceParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO finance_recurring_invoices\n(owner_id, business_id, template_invoice_id, frequency, next_invoice_date, end_date, is_active)\nVALUES (?, ?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.template_invoice_id)
        .bind(&params.frequency)
        .bind(params.next_invoice_date)
        .bind(params.end_date)
        .bind(params.is_active)
        .execute(exec)
        .await
}

/// Params for `GetRecurringInvoice` (Go `GetRecurringInvoiceParams`).
pub struct GetRecurringInvoiceParams {
    pub id: i64,
    pub business_id: i64,
}

/// `GetRecurringInvoice`: `one`.
pub async fn get_recurring_invoice<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetRecurringInvoiceParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, template_invoice_id, frequency, next_invoice_date, end_date, is_active, created_at\nFROM finance_recurring_invoices\nWHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `ListRecurringInvoicesPage` (Go `ListRecurringInvoicesPageParams`).
pub struct ListRecurringInvoicesPageParams {
    pub business_id: i64,
    pub limit: i32,
    pub offset: i32,
}

/// `ListRecurringInvoicesPage`: `many`.
pub async fn list_recurring_invoices_page<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListRecurringInvoicesPageParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, template_invoice_id, frequency, next_invoice_date, end_date, is_active, created_at\nFROM finance_recurring_invoices\nWHERE business_id = ?\nORDER BY next_invoice_date ASC, id ASC\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `RecurringInvoiceExists` (Go `RecurringInvoiceExistsParams`).
pub struct RecurringInvoiceExistsParams {
    pub id: i64,
    pub business_id: i64,
}

/// `RecurringInvoiceExists`: `one`.
pub async fn recurring_invoice_exists<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &RecurringInvoiceExistsParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT EXISTS (\n    SELECT 1 FROM finance_recurring_invoices WHERE id = ? AND business_id = ?\n)")
        .bind(params.id)
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `SetRecurringInvoiceActive` (Go `SetRecurringInvoiceActiveParams`).
pub struct SetRecurringInvoiceActiveParams {
    pub is_active: Option<bool>,
    pub id: i64,
    pub business_id: i64,
}

/// `SetRecurringInvoiceActive`: `execresult`.
pub async fn set_recurring_invoice_active<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &SetRecurringInvoiceActiveParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query(
        "UPDATE finance_recurring_invoices\nSET is_active = ?\nWHERE id = ? AND business_id = ?",
    )
    .bind(params.is_active)
    .bind(params.id)
    .bind(params.business_id)
    .execute(exec)
    .await
}

/// Params for `UpdateRecurringInvoice` (Go `UpdateRecurringInvoiceParams`).
pub struct UpdateRecurringInvoiceParams {
    pub template_invoice_id: Option<i64>,
    pub frequency: String,
    pub next_invoice_date: chrono::NaiveDate,
    pub set_end_date: i64,
    pub end_date: Option<chrono::DateTime<chrono::Utc>>,
    pub is_active: Option<bool>,
    pub id: i64,
    pub business_id: i64,
}

/// `UpdateRecurringInvoice`: `execresult`.
pub async fn update_recurring_invoice<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateRecurringInvoiceParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE finance_recurring_invoices\nSET template_invoice_id = COALESCE(?, template_invoice_id),\n    frequency = ?,\n    next_invoice_date = ?,\n    end_date = CASE WHEN CAST(? AS UNSIGNED) = 1 THEN ? ELSE end_date END,\n    is_active = COALESCE(?, is_active)\nWHERE id = ? AND business_id = ?")
        .bind(params.template_invoice_id)
        .bind(&params.frequency)
        .bind(params.next_invoice_date)
        .bind(params.set_end_date)
        .bind(params.end_date)
        .bind(params.is_active)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}
