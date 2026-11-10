//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CountBillsByPurchaseOrder` (Go `CountBillsByPurchaseOrderParams`).
pub struct CountBillsByPurchaseOrderParams {
    pub business_id: i64,
    pub purchase_order_id: Option<i64>,
}

/// `CountBillsByPurchaseOrder`: `one`.
pub async fn count_bills_by_purchase_order<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountBillsByPurchaseOrderParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query(
        "SELECT COUNT(*) FROM finance_bills WHERE business_id = ? AND purchase_order_id = ?",
    )
    .bind(params.business_id)
    .bind(params.purchase_order_id)
    .fetch_optional(exec)
    .await
}

/// Params for `CountBillsByStatus` (Go `CountBillsByStatusParams`).
pub struct CountBillsByStatusParams {
    pub business_id: i64,
}

/// `CountBillsByStatus`: `many`.
pub async fn count_bills_by_status<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountBillsByStatusParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT status, COUNT(*) AS cnt\nFROM finance_bills\nWHERE business_id = ?\n  AND created_at >= DATE_SUB(UTC_TIMESTAMP(), INTERVAL 12 MONTH)\nGROUP BY status")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `CreateBill` (Go `CreateBillParams`).
pub struct CreateBillParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub supplier_id: Option<i64>,
    pub supplier_name: String,
    pub purchase_order_id: Option<i64>,
    pub amount: String,
    pub currency: Option<String>,
    pub due_date: chrono::NaiveDate,
    pub status: String,
    pub description: Option<String>,
}

/// `CreateBill`: `execresult`.
pub async fn create_bill<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateBillParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO finance_bills (owner_id, business_id, supplier_id, supplier_name, purchase_order_id, amount, currency, due_date, status, description)\nVALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.supplier_id)
        .bind(&params.supplier_name)
        .bind(params.purchase_order_id)
        .bind(&params.amount)
        .bind(&params.currency)
        .bind(params.due_date)
        .bind(&params.status)
        .bind(&params.description)
        .execute(exec)
        .await
}

/// Params for `DeleteBill` (Go `DeleteBillParams`).
pub struct DeleteBillParams {
    pub id: i64,
    pub business_id: i64,
}

/// `DeleteBill`: `execresult`.
pub async fn delete_bill<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteBillParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM finance_bills\nWHERE id = ? AND business_id = ? AND status <> 'paid'")
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `GetBill` (Go `GetBillParams`).
pub struct GetBillParams {
    pub id: i64,
    pub business_id: i64,
}

/// `GetBill`: `one`.
pub async fn get_bill<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetBillParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, supplier_id, COALESCE(supplier_name,'') AS supplier_name, purchase_order_id, amount, currency, due_date, status, description, created_at, updated_at\nFROM finance_bills\nWHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `ListBills` (Go `ListBillsParams`).
pub struct ListBillsParams {
    pub business_id: i64,
}

/// `ListBills`: `many`.
pub async fn list_bills<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListBillsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, supplier_id, COALESCE(supplier_name,'') AS supplier_name, purchase_order_id, amount, currency, due_date, status, description, created_at, updated_at\nFROM finance_bills\nWHERE business_id = ?\nORDER BY due_date ASC, id ASC\nLIMIT 500")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `ListBillsByPurchaseOrder` (Go `ListBillsByPurchaseOrderParams`).
pub struct ListBillsByPurchaseOrderParams {
    pub business_id: i64,
    pub purchase_order_id: Option<i64>,
}

/// `ListBillsByPurchaseOrder`: `many`.
pub async fn list_bills_by_purchase_order<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListBillsByPurchaseOrderParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, supplier_id, COALESCE(supplier_name,'') AS supplier_name, purchase_order_id, amount, currency, due_date, status, description, created_at, updated_at\nFROM finance_bills\nWHERE business_id = ? AND purchase_order_id = ?\nORDER BY due_date ASC, id ASC\nLIMIT 500")
        .bind(params.business_id)
        .bind(params.purchase_order_id)
        .execute(exec)
        .await
}

/// Params for `ListBillsPaginated` (Go `ListBillsPaginatedParams`).
pub struct ListBillsPaginatedParams {
    pub business_id: i64,
    pub limit: i32,
    pub offset: i32,
}

/// `ListBillsPaginated`: `many`.
pub async fn list_bills_paginated<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListBillsPaginatedParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, supplier_id, COALESCE(supplier_name,'') AS supplier_name, purchase_order_id, amount, currency, due_date, status, description, created_at, updated_at\nFROM finance_bills\nWHERE business_id = ?\nORDER BY due_date ASC, id ASC\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `SumBillsByStatus` (Go `SumBillsByStatusParams`).
pub struct SumBillsByStatusParams {
    pub business_id: i64,
    pub status: String,
}

/// `SumBillsByStatus`: `one`.
pub async fn sum_bills_by_status<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &SumBillsByStatusParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COALESCE(SUM(CAST(amount AS DECIMAL(14,2))), 0) AS total\nFROM finance_bills\nWHERE business_id = ? AND status = ?")
        .bind(params.business_id)
        .bind(&params.status)
        .fetch_optional(exec)
        .await
}

/// Params for `UpdateBill` (Go `UpdateBillParams`).
pub struct UpdateBillParams {
    pub supplier_id: Option<i64>,
    pub supplier_name: String,
    pub purchase_order_id: Option<i64>,
    pub amount: String,
    pub currency: Option<String>,
    pub due_date: chrono::NaiveDate,
    pub status: String,
    pub description: Option<String>,
    pub id: i64,
    pub business_id: i64,
    pub current_status: String,
}

/// `UpdateBill`: `execresult`.
pub async fn update_bill<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateBillParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE finance_bills\nSET supplier_id = ?, supplier_name = ?, purchase_order_id = ?, amount = ?, currency = ?, due_date = ?, status = ?, description = ?\nWHERE id = ? AND business_id = ? AND status = ? AND status <> 'paid'")
        .bind(params.supplier_id)
        .bind(&params.supplier_name)
        .bind(params.purchase_order_id)
        .bind(&params.amount)
        .bind(&params.currency)
        .bind(params.due_date)
        .bind(&params.status)
        .bind(&params.description)
        .bind(params.id)
        .bind(params.business_id)
        .bind(&params.current_status)
        .execute(exec)
        .await
}
