//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `GetBranchInventoryItem` (Go `GetBranchInventoryItemParams`).
pub struct GetBranchInventoryItemParams {
    pub source_id: i64,
    pub business_id: i64,
    pub id: i64,
}

/// `GetBranchInventoryItem`: `many`.
pub async fn get_branch_inventory_item<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetBranchInventoryItemParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT\n    p.id AS id,\n    p.id AS product_id,\n    p.name,\n    p.sku,\n    p.unit,\n    COALESCE(sl.on_hand, 0) AS quantity,\n    COALESCE(sl.reorder_point, p.low_stock_threshold) AS reorder_point,\n    COALESCE(sl.reorder_quantity, p.reorder_quantity) AS reorder_quantity,\n    p.low_stock_threshold,\n    COALESCE(sl.updated_at, p.updated_at) AS updated_at\nFROM erp_products p\nLEFT JOIN stock_locations loc\n  ON loc.business_id = p.business_id\n AND loc.source_table = 'pos_locations'\n AND loc.source_id = ?\nLEFT JOIN stock_levels sl\n  ON sl.business_id = p.business_id\n AND sl.location_id = loc.id\n AND sl.product_id = p.id\nWHERE p.business_id = ?\n  AND p.id = ?\n  AND p.is_active = TRUE\n  AND p.track_inventory = TRUE\nLIMIT 1")
        .bind(params.source_id)
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `InventoryProductExists` (Go `InventoryProductExistsParams`).
pub struct InventoryProductExistsParams {
    pub business_id: i64,
    pub id: i64,
}

/// `InventoryProductExists`: `one`.
pub async fn inventory_product_exists<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InventoryProductExistsParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT 1 AS exists_flag\nFROM erp_products\nWHERE business_id = ? AND id = ? AND is_active = TRUE AND track_inventory = TRUE\nLIMIT 1")
        .bind(params.business_id)
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `LastStockLedgerID` (Go `LastStockLedgerIDParams`).
pub struct LastStockLedgerIDParams {
    pub business_id: i64,
    pub location_id: i64,
    pub product_id: i64,
}

/// `LastStockLedgerID`: `one`.
pub async fn last_stock_ledger_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &LastStockLedgerIDParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id FROM stock_ledger\nWHERE business_id = ? AND location_id = ? AND product_id = ?\nORDER BY id DESC\nLIMIT 1")
        .bind(params.business_id)
        .bind(params.location_id)
        .bind(params.product_id)
        .fetch_optional(exec)
        .await
}

/// Params for `LinkStockLedgerTransferIn` (Go `LinkStockLedgerTransferInParams`).
pub struct LinkStockLedgerTransferInParams {
    pub transfer_group: Option<i64>,
    pub reference_id: Option<i64>,
    pub id: i64,
    pub business_id: i64,
}

/// `LinkStockLedgerTransferIn`: `exec`.
pub async fn link_stock_ledger_transfer_in<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &LinkStockLedgerTransferInParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE stock_ledger SET transfer_group = ?, reference_id = ? WHERE id = ? AND business_id = ?")
        .bind(params.transfer_group)
        .bind(params.reference_id)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `ListBranchInventory` (Go `ListBranchInventoryParams`).
pub struct ListBranchInventoryParams {
    pub source_id: i64,
    pub business_id: i64,
}

/// `ListBranchInventory`: `many`.
pub async fn list_branch_inventory<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListBranchInventoryParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT\n    p.id AS id,\n    p.id AS product_id,\n    p.name,\n    p.sku,\n    p.unit,\n    COALESCE(sl.on_hand, 0) AS quantity,\n    COALESCE(sl.reorder_point, p.low_stock_threshold) AS reorder_point,\n    COALESCE(sl.reorder_quantity, p.reorder_quantity) AS reorder_quantity,\n    p.low_stock_threshold,\n    COALESCE(sl.updated_at, p.updated_at) AS updated_at\nFROM erp_products p\nLEFT JOIN stock_locations loc\n  ON loc.business_id = p.business_id\n AND loc.source_table = 'pos_locations'\n AND loc.source_id = ?\nLEFT JOIN stock_levels sl\n  ON sl.business_id = p.business_id\n AND sl.location_id = loc.id\n AND sl.product_id = p.id\nWHERE p.business_id = ?\n  AND p.is_active = TRUE\n  AND p.track_inventory = TRUE\nORDER BY p.name")
        .bind(params.source_id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `ListLowStockBranchInventory` (Go `ListLowStockBranchInventoryParams`).
pub struct ListLowStockBranchInventoryParams {
    pub source_id: i64,
    pub business_id: i64,
}

/// `ListLowStockBranchInventory`: `many`.
pub async fn list_low_stock_branch_inventory<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListLowStockBranchInventoryParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT\n    p.id AS id,\n    p.id AS product_id,\n    p.name,\n    p.sku,\n    p.unit,\n    COALESCE(sl.on_hand, 0) AS quantity,\n    COALESCE(sl.reorder_point, p.low_stock_threshold) AS reorder_point,\n    COALESCE(sl.reorder_quantity, p.reorder_quantity) AS reorder_quantity,\n    p.low_stock_threshold,\n    COALESCE(sl.updated_at, p.updated_at) AS updated_at\nFROM erp_products p\nLEFT JOIN stock_locations loc\n  ON loc.business_id = p.business_id\n AND loc.source_table = 'pos_locations'\n AND loc.source_id = ?\nLEFT JOIN stock_levels sl\n  ON sl.business_id = p.business_id\n AND sl.location_id = loc.id\n AND sl.product_id = p.id\nWHERE p.business_id = ?\n  AND p.is_active = TRUE\n  AND p.track_inventory = TRUE\n  AND COALESCE(sl.reorder_point, p.low_stock_threshold) IS NOT NULL\n  AND COALESCE(sl.on_hand, 0) <= COALESCE(sl.reorder_point, p.low_stock_threshold)\nORDER BY COALESCE(sl.on_hand, 0) ASC, p.name")
        .bind(params.source_id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `ListPOSStockMovements` (Go `ListPOSStockMovementsParams`).
pub struct ListPOSStockMovementsParams {
    pub business_id: i64,
    pub source_id: i64,
    pub product_id: i64,
}

/// `ListPOSStockMovements`: `many`.
pub async fn list_pos_stock_movements<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListPOSStockMovementsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT\n    l.id,\n    l.product_id,\n    l.entry_type,\n    l.quantity_change,\n    l.reference_id,\n    l.reference_type,\n    l.note,\n    l.employee_id,\n    l.created_at\nFROM stock_ledger l\nINNER JOIN stock_locations loc\n  ON loc.id = l.location_id AND loc.business_id = l.business_id\nWHERE l.business_id = ?\n  AND loc.source_table = 'pos_locations'\n  AND loc.source_id = ?\n  AND l.product_id = ?\n  AND l.channel = 'pos'\nORDER BY l.created_at DESC, l.id DESC\nLIMIT 100")
        .bind(params.business_id)
        .bind(params.source_id)
        .bind(params.product_id)
        .execute(exec)
        .await
}

/// Params for `UpdateStockLedgerTransferGroup` (Go `UpdateStockLedgerTransferGroupParams`).
pub struct UpdateStockLedgerTransferGroupParams {
    pub transfer_group: Option<i64>,
    pub id: i64,
    pub business_id: i64,
}

/// `UpdateStockLedgerTransferGroup`: `exec`.
pub async fn update_stock_ledger_transfer_group<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateStockLedgerTransferGroupParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE stock_ledger SET transfer_group = ? WHERE id = ? AND business_id = ?")
        .bind(params.transfer_group)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}
