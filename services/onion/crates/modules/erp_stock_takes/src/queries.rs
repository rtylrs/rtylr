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

/// Params for `GetStockTakeItems` (Go `GetStockTakeItemsParams`).
pub struct GetStockTakeItemsParams {
    pub stock_take_id: i64,
}

/// `GetStockTakeItems`: `many`.
pub async fn get_stock_take_items<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetStockTakeItemsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, stock_take_id, product_id, expected_qty, counted_qty, variance, created_at\nFROM erp_stock_take_items\nWHERE stock_take_id = ?\nORDER BY id ASC")
        .bind(params.stock_take_id)
        .execute(exec)
        .await
}

/// Params for `GetStockTakeItemsByStockTakeIDs` (Go `GetStockTakeItemsByStockTakeIDsParams`).
pub struct GetStockTakeItemsByStockTakeIDsParams {
    pub stock_take_ids: Vec<i64>,
}

/// `GetStockTakeItemsByStockTakeIDs`: `many`.
pub async fn get_stock_take_items_by_stock_take_i_ds<
    'e,
    E: sqlx::Executor<'e, Database = MySql>,
>(
    exec: E,
    params: &GetStockTakeItemsByStockTakeIDsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    let mut sql_template: String = "SELECT id, stock_take_id, product_id, expected_qty, counted_qty, variance, created_at\nFROM erp_stock_take_items\nWHERE stock_take_id IN ({SLICE:stock_take_ids})\nORDER BY id ASC".to_string();

    if params.stock_take_ids.is_empty() {
        sql_template = sql_template.replace("{SLICE:stock_take_ids}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:stock_take_ids}",
            &std::iter::repeat_n("?", params.stock_take_ids.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    let query = sqlx::query(&sql_template);
    let mut query = query;
    for value in &params.stock_take_ids {
        query = query.bind(value);
    }
    query.execute(exec).await
}

/// Params for `GetStockTakeStatusForUpdate` (Go `GetStockTakeStatusForUpdateParams`).
pub struct GetStockTakeStatusForUpdateParams {
    pub business_id: i64,
    pub id: i64,
}

/// `GetStockTakeStatusForUpdate`: `one`.
pub async fn get_stock_take_status_for_update<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetStockTakeStatusForUpdateParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT warehouse_id, status\nFROM erp_stock_takes\nWHERE business_id = ? AND id = ?\nFOR UPDATE")
        .bind(params.business_id)
        .bind(params.id)
        .fetch_optional(exec)
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

/// Params for `InsertInventoryTransactionWithRef` (Go `InsertInventoryTransactionWithRefParams`).
pub struct InsertInventoryTransactionWithRefParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub product_id: i64,
    pub r#type: String,
    pub quantity_change: String,
    pub quantity_before: String,
    pub quantity_after: String,
    pub reference_id: Option<i64>,
    pub reference_type: Option<String>,
    pub created_by: Option<i64>,
    pub notes: Option<String>,
}

/// `InsertInventoryTransactionWithRef`: `execresult`.
pub async fn insert_inventory_transaction_with_ref<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsertInventoryTransactionWithRefParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO erp_inventory_transactions\n(owner_id, business_id, product_id, type, quantity_change, quantity_before, quantity_after,\n reference_id, reference_type, created_by, notes)\nVALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.product_id)
        .bind(&params.r#type)
        .bind(&params.quantity_change)
        .bind(&params.quantity_before)
        .bind(&params.quantity_after)
        .bind(params.reference_id)
        .bind(&params.reference_type)
        .bind(params.created_by)
        .bind(&params.notes)
        .execute(exec)
        .await
}

/// Params for `InsertStockTake` (Go `InsertStockTakeParams`).
pub struct InsertStockTakeParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub warehouse_id: Option<i64>,
    pub notes: Option<String>,
}

/// `InsertStockTake`: `execresult`.
pub async fn insert_stock_take<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsertStockTakeParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO erp_stock_takes (owner_id, business_id, warehouse_id, status, notes)\nVALUES (?, ?, ?, 'draft', ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.warehouse_id)
        .bind(&params.notes)
        .execute(exec)
        .await
}

/// Params for `InsertStockTakeItem` (Go `InsertStockTakeItemParams`).
pub struct InsertStockTakeItemParams {
    pub stock_take_id: i64,
    pub product_id: i64,
    pub expected_qty: String,
    pub counted_qty: Option<String>,
    pub variance: Option<String>,
}

/// `InsertStockTakeItem`: `execresult`.
pub async fn insert_stock_take_item<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsertStockTakeItemParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO erp_stock_take_items (stock_take_id, product_id, expected_qty, counted_qty, variance)\nVALUES (?, ?, ?, ?, ?)")
        .bind(params.stock_take_id)
        .bind(params.product_id)
        .bind(&params.expected_qty)
        .bind(&params.counted_qty)
        .bind(&params.variance)
        .execute(exec)
        .await
}

/// Params for `ListStockTakes` (Go `ListStockTakesParams`).
pub struct ListStockTakesParams {
    pub business_id: i64,
}

/// `ListStockTakes`: `many`.
pub async fn list_stock_takes<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListStockTakesParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, warehouse_id, status, notes, started_at, completed_at, created_at\nFROM erp_stock_takes\nWHERE business_id = ?\nORDER BY created_at DESC")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `ListStockTakesPaged` (Go `ListStockTakesPagedParams`).
pub struct ListStockTakesPagedParams {
    pub business_id: i64,
    pub limit: i32,
    pub offset: i32,
}

/// `ListStockTakesPaged`: `many`.
pub async fn list_stock_takes_paged<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListStockTakesPagedParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, warehouse_id, status, notes, started_at, completed_at, created_at\nFROM erp_stock_takes\nWHERE business_id = ?\nORDER BY created_at DESC, id DESC\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `UpdateStockTakeItemCount` (Go `UpdateStockTakeItemCountParams`).
pub struct UpdateStockTakeItemCountParams {
    pub counted_qty: Option<String>,
    pub variance: Option<String>,
    pub stock_take_id: i64,
    pub product_id: i64,
}

/// `UpdateStockTakeItemCount`: `execrows`.
pub async fn update_stock_take_item_count<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateStockTakeItemCountParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE erp_stock_take_items\nSET counted_qty = ?, variance = ?\nWHERE stock_take_id = ? AND product_id = ?")
        .bind(&params.counted_qty)
        .bind(&params.variance)
        .bind(params.stock_take_id)
        .bind(params.product_id)
        .execute(exec)
        .await
}

/// Params for `UpdateStockTakeStatus` (Go `UpdateStockTakeStatusParams`).
pub struct UpdateStockTakeStatusParams {
    pub status: String,
    pub notes: Option<String>,
    pub business_id: i64,
    pub id: i64,
}

/// `UpdateStockTakeStatus`: `exec`.
pub async fn update_stock_take_status<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateStockTakeStatusParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query(
        "UPDATE erp_stock_takes\nSET status = ?, notes = ?\nWHERE business_id = ? AND id = ?",
    )
    .bind(&params.status)
    .bind(&params.notes)
    .bind(params.business_id)
    .bind(params.id)
    .execute(exec)
    .await
}

/// Params for `UpdateStockTakeStatusCompleted` (Go `UpdateStockTakeStatusCompletedParams`).
pub struct UpdateStockTakeStatusCompletedParams {
    pub status: String,
    pub notes: Option<String>,
    pub business_id: i64,
    pub id: i64,
}

/// `UpdateStockTakeStatusCompleted`: `exec`.
pub async fn update_stock_take_status_completed<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateStockTakeStatusCompletedParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE erp_stock_takes\nSET status = ?, notes = ?, completed_at = NOW()\nWHERE business_id = ? AND id = ?")
        .bind(&params.status)
        .bind(&params.notes)
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `UpdateStockTakeStatusInProgress` (Go `UpdateStockTakeStatusInProgressParams`).
pub struct UpdateStockTakeStatusInProgressParams {
    pub status: String,
    pub notes: Option<String>,
    pub business_id: i64,
    pub id: i64,
}

/// `UpdateStockTakeStatusInProgress`: `exec`.
pub async fn update_stock_take_status_in_progress<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateStockTakeStatusInProgressParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE erp_stock_takes\nSET status = ?, notes = ?, started_at = NOW()\nWHERE business_id = ? AND id = ?")
        .bind(&params.status)
        .bind(&params.notes)
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}
