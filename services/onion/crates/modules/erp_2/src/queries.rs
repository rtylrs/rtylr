//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `EnsureStockLevelRow` (Go `EnsureStockLevelRowParams`).
pub struct EnsureStockLevelRowParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub location_id: i64,
    pub product_id: i64,
}

/// `EnsureStockLevelRow`: `exec`.
pub async fn ensure_stock_level_row<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &EnsureStockLevelRowParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO stock_levels (owner_id, business_id, location_id, product_id, on_hand)\n VALUES (?, ?, ?, ?, 0) ON DUPLICATE KEY UPDATE on_hand = on_hand")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.location_id)
        .bind(params.product_id)
        .execute(exec)
        .await
}

/// Params for `GetDefaultWarehouseLocationID` (Go `GetDefaultWarehouseLocationIDParams`).
pub struct GetDefaultWarehouseLocationIDParams {
    pub owner_id: i64,
    pub business_id: i64,
}

/// `GetDefaultWarehouseLocationID`: `one`.
pub async fn get_default_warehouse_location_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetDefaultWarehouseLocationIDParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id FROM stock_locations\n WHERE owner_id = ? AND business_id = ? AND kind = 'warehouse' AND source_table = 'native' AND source_id = 0\n LIMIT 1")
        .bind(params.owner_id)
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetPOSBranchLocationID` (Go `GetPOSBranchLocationIDParams`).
pub struct GetPOSBranchLocationIDParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub source_id: i64,
}

/// `GetPOSBranchLocationID`: `one`.
pub async fn get_pos_branch_location_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetPOSBranchLocationIDParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id FROM stock_locations\n WHERE owner_id = ? AND business_id = ? AND source_table = 'pos_locations' AND source_id = ? LIMIT 1")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.source_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetWarehouseBusinessID` (Go `GetWarehouseBusinessIDParams`).
pub struct GetWarehouseBusinessIDParams {
    pub id: i64,
}

/// `GetWarehouseBusinessID`: `one`.
pub async fn get_warehouse_business_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetWarehouseBusinessIDParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT business_id FROM erp_warehouses WHERE id = ?")
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetWarehouseLocationID` (Go `GetWarehouseLocationIDParams`).
pub struct GetWarehouseLocationIDParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub source_id: i64,
}

/// `GetWarehouseLocationID`: `one`.
pub async fn get_warehouse_location_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetWarehouseLocationIDParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id FROM stock_locations\n WHERE owner_id = ? AND business_id = ? AND source_table = 'erp_warehouses' AND source_id = ? LIMIT 1")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.source_id)
        .fetch_optional(exec)
        .await
}

/// Params for `InsertDefaultWarehouseLocation` (Go `InsertDefaultWarehouseLocationParams`).
pub struct InsertDefaultWarehouseLocationParams {
    pub owner_id: i64,
    pub business_id: i64,
}

/// `InsertDefaultWarehouseLocation`: `exec`.
pub async fn insert_default_warehouse_location<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsertDefaultWarehouseLocationParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO stock_locations (owner_id, business_id, kind, source_table, source_id, name, is_default, is_active)\n VALUES (?, ?, 'warehouse', 'native', 0, 'Default Warehouse', TRUE, TRUE)\n ON DUPLICATE KEY UPDATE id = id")
        .bind(params.owner_id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `InsertPOSBranchLocation` (Go `InsertPOSBranchLocationParams`).
pub struct InsertPOSBranchLocationParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub source_id: i64,
}

/// `InsertPOSBranchLocation`: `exec`.
pub async fn insert_pos_branch_location<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsertPOSBranchLocationParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO stock_locations (owner_id, business_id, kind, source_table, source_id, name, is_default, is_active)\n VALUES (?, ?, 'pos_branch', 'pos_locations', ?, 'POS Branch', FALSE, TRUE)\n ON DUPLICATE KEY UPDATE source_id = VALUES(source_id)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.source_id)
        .execute(exec)
        .await
}

/// Params for `InsertStockLedgerEntry` (Go `InsertStockLedgerEntryParams`).
pub struct InsertStockLedgerEntryParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub location_id: i64,
    pub product_id: i64,
    pub entry_type: String,
    pub quantity_change: String,
    pub quantity_before: String,
    pub quantity_after: String,
    pub reference_id: Option<i64>,
    pub reference_type: Option<String>,
    pub transfer_group: Option<i64>,
    pub channel: String,
    pub note: Option<String>,
    pub employee_id: Option<i64>,
    pub created_by: Option<i64>,
}

/// `InsertStockLedgerEntry`: `execresult`.
pub async fn insert_stock_ledger_entry<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsertStockLedgerEntryParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO stock_ledger\n (owner_id, business_id, location_id, product_id, entry_type, quantity_change, quantity_before, quantity_after,\n  reference_id, reference_type, transfer_group, channel, note, employee_id, created_by)\n VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.location_id)
        .bind(params.product_id)
        .bind(&params.entry_type)
        .bind(&params.quantity_change)
        .bind(&params.quantity_before)
        .bind(&params.quantity_after)
        .bind(params.reference_id)
        .bind(&params.reference_type)
        .bind(params.transfer_group)
        .bind(&params.channel)
        .bind(&params.note)
        .bind(params.employee_id)
        .bind(params.created_by)
        .execute(exec)
        .await
}

/// Params for `InsertWarehouseLocation` (Go `InsertWarehouseLocationParams`).
pub struct InsertWarehouseLocationParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub id: i64,
}

/// `InsertWarehouseLocation`: `exec`.
pub async fn insert_warehouse_location<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsertWarehouseLocationParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO stock_locations (owner_id, business_id, kind, source_table, source_id, name, is_default, is_active)\n SELECT ?, ?, 'warehouse', 'erp_warehouses', w.id, w.name, FALSE, w.is_active FROM erp_warehouses w WHERE w.id = ?\n ON DUPLICATE KEY UPDATE source_id = VALUES(source_id)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `LockStockLevelOnHand` (Go `LockStockLevelOnHandParams`).
pub struct LockStockLevelOnHandParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub location_id: i64,
    pub product_id: i64,
}

/// `LockStockLevelOnHand`: `one`.
pub async fn lock_stock_level_on_hand<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &LockStockLevelOnHandParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT on_hand FROM stock_levels\n WHERE owner_id = ? AND business_id = ? AND location_id = ? AND product_id = ? FOR UPDATE")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.location_id)
        .bind(params.product_id)
        .fetch_optional(exec)
        .await
}

/// Params for `SyncProductStockScalar` (Go `SyncProductStockScalarParams`).
pub struct SyncProductStockScalarParams {
    pub stock_quantity: String,
    pub business_id: i64,
    pub id: i64,
}

/// `SyncProductStockScalar`: `exec`.
pub async fn sync_product_stock_scalar<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &SyncProductStockScalarParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE erp_products SET stock_quantity = stock_quantity + ?\n WHERE business_id = ? AND id = ?")
        .bind(&params.stock_quantity)
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `UpdateStockLevelOnHand` (Go `UpdateStockLevelOnHandParams`).
pub struct UpdateStockLevelOnHandParams {
    pub on_hand: String,
    pub owner_id: i64,
    pub business_id: i64,
    pub location_id: i64,
    pub product_id: i64,
}

/// `UpdateStockLevelOnHand`: `exec`.
pub async fn update_stock_level_on_hand<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateStockLevelOnHandParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE stock_levels SET on_hand = ?\n WHERE owner_id = ? AND business_id = ? AND location_id = ? AND product_id = ?")
        .bind(&params.on_hand)
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.location_id)
        .bind(params.product_id)
        .execute(exec)
        .await
}
