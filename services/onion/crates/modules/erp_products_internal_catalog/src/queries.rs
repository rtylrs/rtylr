//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `ListMenuItemsWithDetails` (Go `ListMenuItemsWithDetailsParams`).
pub struct ListMenuItemsWithDetailsParams {
    pub business_id: i64,
    pub business_id_2: i64,
}

/// `ListMenuItemsWithDetails`: `many`.
pub async fn list_menu_items_with_details<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListMenuItemsWithDetailsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT\n    p.id,\n    p.owner_id,\n    p.business_id,\n    p.category_id,\n    COALESCE(c.name, 'Other') AS category_name,\n    p.name,\n    p.sku,\n    p.barcode,\n    p.sale_price,\n    p.cost_price,\n    p.tax_rate,\n    p.description,\n    COALESCE(img.url, '') AS image_url,\n    COALESCE(pst.total_quantity, p.stock_quantity) AS stock_quantity,\n    p.is_active,\n    p.created_at,\n    p.updated_at\nFROM erp_products p\nLEFT JOIN erp_categories c\n  ON c.id = p.category_id\n AND c.business_id = p.business_id\nLEFT JOIN product_stock_total pst\n  ON pst.business_id = p.business_id\n AND pst.product_id = p.id\nLEFT JOIN (\n    SELECT business_id, product_id, url\n    FROM (\n        SELECT\n            epi.business_id,\n            epi.product_id,\n            epi.url,\n            ROW_NUMBER() OVER (\n                PARTITION BY epi.business_id, epi.product_id\n                ORDER BY epi.is_primary DESC, epi.position ASC, epi.id ASC\n            ) AS image_rank\n        FROM erp_product_images epi\n        WHERE epi.business_id = ?\n    ) ranked_product_images\n    WHERE image_rank = 1\n) img\n  ON img.business_id = p.business_id\n AND img.product_id = p.id\nWHERE p.business_id = ?\n  AND p.is_active = TRUE\nORDER BY c.name, p.name, p.id")
        .bind(params.business_id)
        .bind(params.business_id_2)
        .execute(exec)
        .await
}

/// Params for `QuickCreateMenuItem` (Go `QuickCreateMenuItemParams`).
pub struct QuickCreateMenuItemParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub name: String,
    pub sku: Option<String>,
    pub sale_price: String,
}

/// `QuickCreateMenuItem`: `execresult`.
pub async fn quick_create_menu_item<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &QuickCreateMenuItemParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO erp_products (owner_id, business_id, name, sku, sale_price, is_active)\nVALUES (?, ?, ?, ?, ?, TRUE)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(&params.name)
        .bind(&params.sku)
        .bind(&params.sale_price)
        .execute(exec)
        .await
}

/// Params for `UpdateMenuItem` (Go `UpdateMenuItemParams`).
pub struct UpdateMenuItemParams {
    pub name: String,
    pub sale_price: String,
    pub is_active: Option<bool>,
    pub business_id: i64,
    pub id: i64,
}

/// `UpdateMenuItem`: `execresult`.
pub async fn update_menu_item<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateMenuItemParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE erp_products\nSET name = COALESCE(NULLIF(?, ''), name),\n    sale_price = ?,\n    is_active = ?,\n    updated_at = CURRENT_TIMESTAMP\nWHERE business_id = ? AND id = ?")
        .bind(&params.name)
        .bind(&params.sale_price)
        .bind(params.is_active)
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}
