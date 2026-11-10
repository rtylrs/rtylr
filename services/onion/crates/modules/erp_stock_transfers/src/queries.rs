//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `GetProductReference` (Go `GetProductReferenceParams`).
pub struct GetProductReferenceParams {
    pub business_id: i64,
    pub id: i64,
}

/// `GetProductReference`: `one`.
pub async fn get_product_reference<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetProductReferenceParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id FROM erp_products WHERE business_id = ? AND id = ?")
        .bind(params.business_id)
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetStockTransferByID` (Go `GetStockTransferByIDParams`).
pub struct GetStockTransferByIDParams {
    pub business_id: i64,
    pub id: i64,
}

/// `GetStockTransferByID`: `one`.
pub async fn get_stock_transfer_by_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetStockTransferByIDParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, from_warehouse_id, to_warehouse_id, status, notes, created_at, updated_at\nFROM erp_stock_transfers\nWHERE business_id = ? AND id = ?\nORDER BY id ASC\nLIMIT 1")
        .bind(params.business_id)
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetStockTransferForUpdate` (Go `GetStockTransferForUpdateParams`).
pub struct GetStockTransferForUpdateParams {
    pub business_id: i64,
    pub id: i64,
}

/// `GetStockTransferForUpdate`: `one`.
pub async fn get_stock_transfer_for_update<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetStockTransferForUpdateParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT from_warehouse_id, to_warehouse_id, status\nFROM erp_stock_transfers\nWHERE business_id = ? AND id = ?\nFOR UPDATE")
        .bind(params.business_id)
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetTransferItems` (Go `GetTransferItemsParams`).
pub struct GetTransferItemsParams {
    pub transfer_id: i64,
}

/// `GetTransferItems`: `many`.
pub async fn get_transfer_items<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetTransferItemsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, transfer_id, product_id, quantity, created_at\nFROM erp_stock_transfer_items\nWHERE transfer_id = ?\nORDER BY id ASC")
        .bind(params.transfer_id)
        .execute(exec)
        .await
}

/// Params for `GetWarehouseReference` (Go `GetWarehouseReferenceParams`).
pub struct GetWarehouseReferenceParams {
    pub business_id: i64,
    pub id: i64,
}

/// `GetWarehouseReference`: `one`.
pub async fn get_warehouse_reference<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetWarehouseReferenceParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id FROM erp_warehouses WHERE business_id = ? AND id = ?")
        .bind(params.business_id)
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `InsertStockTransfer` (Go `InsertStockTransferParams`).
pub struct InsertStockTransferParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub from_warehouse_id: i64,
    pub to_warehouse_id: i64,
    pub notes: Option<String>,
}

/// `InsertStockTransfer`: `execresult`.
pub async fn insert_stock_transfer<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsertStockTransferParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO erp_stock_transfers (owner_id, business_id, from_warehouse_id, to_warehouse_id, status, notes)\nVALUES (?, ?, ?, ?, 'draft', ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.from_warehouse_id)
        .bind(params.to_warehouse_id)
        .bind(&params.notes)
        .execute(exec)
        .await
}

/// Params for `InsertStockTransferItem` (Go `InsertStockTransferItemParams`).
pub struct InsertStockTransferItemParams {
    pub transfer_id: i64,
    pub product_id: i64,
    pub quantity: String,
}

/// `InsertStockTransferItem`: `execresult`.
pub async fn insert_stock_transfer_item<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsertStockTransferItemParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO erp_stock_transfer_items (transfer_id, product_id, quantity)\nVALUES (?, ?, ?)")
        .bind(params.transfer_id)
        .bind(params.product_id)
        .bind(&params.quantity)
        .execute(exec)
        .await
}

/// Params for `ListStockTransfers` (Go `ListStockTransfersParams`).
pub struct ListStockTransfersParams {
    pub business_id: i64,
}

/// `ListStockTransfers`: `many`.
pub async fn list_stock_transfers<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListStockTransfersParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, from_warehouse_id, to_warehouse_id, status, notes, created_at, updated_at\nFROM erp_stock_transfers\nWHERE business_id = ?\nORDER BY created_at DESC, id DESC\nLIMIT 200")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `ListStockTransfersPaged` (Go `ListStockTransfersPagedParams`).
pub struct ListStockTransfersPagedParams {
    pub business_id: i64,
    pub limit: i32,
    pub offset: i32,
}

/// `ListStockTransfersPaged`: `many`.
pub async fn list_stock_transfers_paged<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListStockTransfersPagedParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, from_warehouse_id, to_warehouse_id, status, notes, created_at, updated_at\nFROM erp_stock_transfers\nWHERE business_id = ?\nORDER BY created_at DESC, id DESC\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `ListTransferItemsByTransferIDs` (Go `ListTransferItemsByTransferIDsParams`).
pub struct ListTransferItemsByTransferIDsParams {
    pub transfer_ids: Vec<i64>,
}

/// `ListTransferItemsByTransferIDs`: `many`.
pub async fn list_transfer_items_by_transfer_i_ds<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListTransferItemsByTransferIDsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    let mut sql_template: String = "SELECT id, transfer_id, product_id, quantity, created_at\nFROM erp_stock_transfer_items\nWHERE transfer_id IN ({SLICE:transfer_ids})\nORDER BY id ASC".to_string();

    if params.transfer_ids.is_empty() {
        sql_template = sql_template.replace("{SLICE:transfer_ids}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:transfer_ids}",
            &std::iter::repeat_n("?", params.transfer_ids.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    let query = sqlx::query(&sql_template);
    let mut query = query;
    for value in &params.transfer_ids {
        query = query.bind(value);
    }
    query.execute(exec).await
}

/// Params for `UpdateStockTransferStatus` (Go `UpdateStockTransferStatusParams`).
pub struct UpdateStockTransferStatusParams {
    pub status: String,
    pub business_id: i64,
    pub id: i64,
}

/// `UpdateStockTransferStatus`: `execresult`.
pub async fn update_stock_transfer_status<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateStockTransferStatusParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE erp_stock_transfers\nSET status = ?\nWHERE business_id = ? AND id = ?\n  AND (status IS NULL OR status <> 'received')")
        .bind(&params.status)
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}
