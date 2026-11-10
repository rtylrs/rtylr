//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CountLowStockProducts` (Go `CountLowStockProductsParams`).
pub struct CountLowStockProductsParams {
    pub business_id: i64,
}

/// `CountLowStockProducts`: `one`.
pub async fn count_low_stock_products<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountLowStockProductsParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) AS total\nFROM erp_products\nWHERE business_id = ? AND track_inventory = TRUE AND is_active = TRUE\n  AND low_stock_threshold IS NOT NULL\n  AND stock_quantity <= low_stock_threshold")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetInventoryReport` (Go `GetInventoryReportParams`).
pub struct GetInventoryReportParams {
    pub business_id: i64,
    pub limit: i32,
    pub offset: i32,
}

/// `GetInventoryReport`: `many`.
pub async fn get_inventory_report<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetInventoryReportParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, business_id, name, sku, unit, stock_quantity, low_stock_threshold,\n       reorder_quantity, cost_price, sale_price,\n       CAST(stock_quantity * cost_price AS DECIMAL(18,4)) AS stock_value,\n       CASE WHEN low_stock_threshold IS NOT NULL AND stock_quantity <= low_stock_threshold THEN 1 ELSE 0 END AS is_low_stock\nFROM erp_products\nWHERE business_id = ? AND track_inventory = TRUE AND is_active = TRUE\nORDER BY name ASC, id ASC\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `GetInventoryReportSummary` (Go `GetInventoryReportSummaryParams`).
pub struct GetInventoryReportSummaryParams {
    pub business_id: i64,
}

/// `GetInventoryReportSummary`: `one`.
pub async fn get_inventory_report_summary<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetInventoryReportSummaryParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) AS total,\n       COALESCE(SUM(CAST(stock_quantity * cost_price AS DECIMAL(18,4))), 0) AS total_value\nFROM erp_products\nWHERE business_id = ? AND track_inventory = TRUE AND is_active = TRUE")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetLowStockProducts` (Go `GetLowStockProductsParams`).
pub struct GetLowStockProductsParams {
    pub business_id: i64,
    pub limit: i32,
    pub offset: i32,
}

/// `GetLowStockProducts`: `many`.
pub async fn get_low_stock_products<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetLowStockProductsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, business_id, name, sku, unit, stock_quantity, low_stock_threshold,\n       reorder_quantity, cost_price, sale_price,\n       CAST(stock_quantity * cost_price AS DECIMAL(18,4)) AS stock_value,\n       1 AS is_low_stock\nFROM erp_products\nWHERE business_id = ? AND track_inventory = TRUE AND is_active = TRUE\n  AND low_stock_threshold IS NOT NULL\n  AND stock_quantity <= low_stock_threshold\nORDER BY stock_quantity ASC, id ASC\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `GetValuation` (Go `GetValuationParams`).
pub struct GetValuationParams {
    pub business_id: i64,
    pub business_id_2: i64,
    pub limit: i32,
    pub offset: i32,
}

/// `GetValuation`: `many`.
pub async fn get_valuation<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetValuationParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT\n    COALESCE(c.name, 'Uncategorized') AS category_name,\n    l.name AS warehouse_name,\n    COUNT(DISTINCT p.id) AS total_items,\n    SUM(COALESCE(cl.total_cost, 0)) AS total_value,\n    SUM(COALESCE(cl.total_cost, 0)) AS total_cost\nFROM stock_levels sl\nJOIN stock_locations l\n  ON l.id = sl.location_id AND l.owner_id = sl.owner_id AND l.is_active = TRUE\nJOIN erp_products p\n  ON p.id = sl.product_id AND p.business_id = sl.business_id\nLEFT JOIN erp_categories c ON c.id = p.category_id AND c.business_id = p.business_id\nLEFT JOIN (\n    SELECT owner_id, business_id, product_id, location_id,\n           SUM(quantity_remaining * unit_cost) AS total_cost\n    FROM inventory_cost_layers\n    WHERE inventory_cost_layers.business_id = ?\n      AND quantity_remaining > 0\n    GROUP BY owner_id, business_id, product_id, location_id\n) cl ON cl.owner_id = sl.owner_id\n    AND cl.business_id = sl.business_id\n    AND cl.product_id = sl.product_id\n    AND cl.location_id = sl.location_id\nWHERE sl.business_id = ? AND p.is_active = TRUE AND p.track_inventory = TRUE\n  AND sl.on_hand > 0\nGROUP BY c.name, l.name\nORDER BY total_value DESC, category_name ASC, warehouse_name ASC\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(params.business_id_2)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `GetValuationSummary` (Go `GetValuationSummaryParams`).
pub struct GetValuationSummaryParams {
    pub business_id: i64,
    pub business_id_2: i64,
}

/// `GetValuationSummary`: `one`.
pub async fn get_valuation_summary<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetValuationSummaryParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) AS total,\n       COALESCE(SUM(total_value), 0) AS total_value\nFROM (\n    SELECT c.name AS category_name,\n           l.name AS warehouse_name,\n           SUM(COALESCE(cl.total_cost, 0)) AS total_value\n    FROM stock_levels sl\n    JOIN stock_locations l\n      ON l.id = sl.location_id AND l.owner_id = sl.owner_id AND l.is_active = TRUE\n    JOIN erp_products p\n      ON p.id = sl.product_id AND p.business_id = sl.business_id\n    LEFT JOIN erp_categories c ON c.id = p.category_id AND c.business_id = p.business_id\n    LEFT JOIN (\n        SELECT owner_id, business_id, product_id, location_id,\n               SUM(quantity_remaining * unit_cost) AS total_cost\n        FROM inventory_cost_layers\n        WHERE inventory_cost_layers.business_id = ?\n          AND quantity_remaining > 0\n        GROUP BY owner_id, business_id, product_id, location_id\n    ) cl ON cl.owner_id = sl.owner_id\n        AND cl.business_id = sl.business_id\n        AND cl.product_id = sl.product_id\n        AND cl.location_id = sl.location_id\n    WHERE sl.business_id = ? AND p.is_active = TRUE AND p.track_inventory = TRUE\n      AND sl.on_hand > 0\n    GROUP BY c.name, l.name\n) AS valuation_groups")
        .bind(params.business_id)
        .bind(params.business_id_2)
        .fetch_optional(exec)
        .await
}

/// Params for `InsertInventoryTransaction` (Go `InsertInventoryTransactionParams`).
pub struct InsertInventoryTransactionParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub product_id: i64,
    pub r#type: String,
    pub quantity_change: String,
    pub quantity_before: String,
    pub quantity_after: String,
    pub notes: Option<String>,
}

/// `InsertInventoryTransaction`: `execresult`.
pub async fn insert_inventory_transaction<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsertInventoryTransactionParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO erp_inventory_transactions\n(owner_id, business_id, product_id, type, quantity_change, quantity_before, quantity_after, notes)\nVALUES (?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.product_id)
        .bind(&params.r#type)
        .bind(&params.quantity_change)
        .bind(&params.quantity_before)
        .bind(&params.quantity_after)
        .bind(&params.notes)
        .execute(exec)
        .await
}
