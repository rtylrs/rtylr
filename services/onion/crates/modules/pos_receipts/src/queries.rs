//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `GetReceiptByID` (Go `GetReceiptByIDParams`).
pub struct GetReceiptByIDParams {
    pub id: i64,
    pub business_id: i64,
}

/// `GetReceiptByID`: `one`.
pub async fn get_receipt_by_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetReceiptByIDParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, order_number, customer_name, total, payment_method, paid_at, business_id\nFROM pos_orders\nWHERE id = ? AND business_id = ? AND payment_status = 'paid'")
        .bind(params.id)
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `ListReceipts` (Go `ListReceiptsParams`).
pub struct ListReceiptsParams {
    pub business_id: i64,
}

/// `ListReceipts`: `many`.
pub async fn list_receipts<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListReceiptsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, order_number, customer_name, total, payment_method, paid_at, business_id\nFROM pos_orders\nWHERE business_id = ? AND payment_status = 'paid'\nORDER BY paid_at DESC, id DESC\nLIMIT 500")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `ListReceiptsWithDateFrom` (Go `ListReceiptsWithDateFromParams`).
pub struct ListReceiptsWithDateFromParams {
    pub business_id: i64,
    pub paid_at: Option<chrono::DateTime<chrono::Utc>>,
}

/// `ListReceiptsWithDateFrom`: `many`.
pub async fn list_receipts_with_date_from<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListReceiptsWithDateFromParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, order_number, customer_name, total, payment_method, paid_at, business_id\nFROM pos_orders\nWHERE business_id = ? AND payment_status = 'paid' AND paid_at >= ?\nORDER BY paid_at DESC, id DESC\nLIMIT 500")
        .bind(params.business_id)
        .bind(params.paid_at)
        .execute(exec)
        .await
}

/// Params for `ListReceiptsWithDateRange` (Go `ListReceiptsWithDateRangeParams`).
pub struct ListReceiptsWithDateRangeParams {
    pub business_id: i64,
    pub paid_at: Option<chrono::DateTime<chrono::Utc>>,
    pub paid_at_2: Option<chrono::DateTime<chrono::Utc>>,
}

/// `ListReceiptsWithDateRange`: `many`.
pub async fn list_receipts_with_date_range<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListReceiptsWithDateRangeParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, order_number, customer_name, total, payment_method, paid_at, business_id\nFROM pos_orders\nWHERE business_id = ? AND payment_status = 'paid' AND paid_at >= ? AND paid_at < ?\nORDER BY paid_at DESC, id DESC\nLIMIT 500")
        .bind(params.business_id)
        .bind(params.paid_at)
        .bind(params.paid_at_2)
        .execute(exec)
        .await
}

/// Params for `ListReceiptsWithDateTo` (Go `ListReceiptsWithDateToParams`).
pub struct ListReceiptsWithDateToParams {
    pub business_id: i64,
    pub paid_at: Option<chrono::DateTime<chrono::Utc>>,
}

/// `ListReceiptsWithDateTo`: `many`.
pub async fn list_receipts_with_date_to<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListReceiptsWithDateToParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, order_number, customer_name, total, payment_method, paid_at, business_id\nFROM pos_orders\nWHERE business_id = ? AND payment_status = 'paid' AND paid_at < ?\nORDER BY paid_at DESC, id DESC\nLIMIT 500")
        .bind(params.business_id)
        .bind(params.paid_at)
        .execute(exec)
        .await
}
