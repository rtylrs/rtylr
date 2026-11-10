//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CountActiveProducts` (Go `CountActiveProductsParams`).
pub struct CountActiveProductsParams {
    pub business_id: i64,
}

/// `CountActiveProducts`: `one`.
pub async fn count_active_products<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountActiveProductsParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM erp_products WHERE business_id = ? AND is_active = TRUE")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CountActiveSuppliers` (Go `CountActiveSuppliersParams`).
pub struct CountActiveSuppliersParams {
    pub business_id: i64,
}

/// `CountActiveSuppliers`: `one`.
pub async fn count_active_suppliers<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountActiveSuppliersParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM erp_suppliers WHERE business_id = ? AND is_active = TRUE")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CountAllProducts` (Go `CountAllProductsParams`).
pub struct CountAllProductsParams {
    pub business_id: i64,
}

/// `CountAllProducts`: `one`.
pub async fn count_all_products<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountAllProductsParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM erp_products WHERE business_id = ?")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CountLowStockProducts` (Go `CountLowStockProductsParams`).
pub struct CountLowStockProductsParams {
    pub business_id: i64,
}

/// `CountLowStockProducts`: `one`.
pub async fn count_low_stock_products<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountLowStockProductsParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM erp_products\nWHERE business_id = ? AND is_active = TRUE\n  AND track_inventory = TRUE\n  AND low_stock_threshold IS NOT NULL\n  AND stock_quantity <= low_stock_threshold")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CountPendingPurchaseOrders` (Go `CountPendingPurchaseOrdersParams`).
pub struct CountPendingPurchaseOrdersParams {
    pub business_id: i64,
}

/// `CountPendingPurchaseOrders`: `one`.
pub async fn count_pending_purchase_orders<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountPendingPurchaseOrdersParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM erp_purchase_orders\nWHERE business_id = ? AND status IN ('draft','sent','partial')")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CountProductsByCategory` (Go `CountProductsByCategoryParams`).
pub struct CountProductsByCategoryParams {
    pub business_id: i64,
    pub category_id: Option<i64>,
}

/// `CountProductsByCategory`: `one`.
pub async fn count_products_by_category<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountProductsByCategoryParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM erp_products WHERE business_id = ? AND category_id = ?")
        .bind(params.business_id)
        .bind(params.category_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CreateProduct` (Go `CreateProductParams`).
pub struct CreateProductParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub category_id: Option<i64>,
    pub supplier_id: Option<i64>,
    pub brand_id: Option<i64>,
    pub name: String,
    pub sku: Option<String>,
    pub barcode: Option<String>,
    pub description: Option<String>,
    pub unit: Option<String>,
    pub cost_price: String,
    pub sale_price: String,
    pub compare_at_price: String,
    pub tax_rate: String,
    pub currency: String,
    pub low_stock_threshold: Option<String>,
    pub reorder_quantity: Option<String>,
    pub allow_backorder: bool,
    pub lead_time_days: i32,
    pub weight: String,
    pub weight_unit: String,
    pub length: String,
    pub width: String,
    pub height: String,
    pub dimension_unit: String,
    pub country_of_origin: Option<String>,
    pub hs_code: Option<String>,
    pub is_active: Option<bool>,
    pub track_inventory: Option<bool>,
}

/// `CreateProduct`: `execresult`.
pub async fn create_product<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateProductParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO erp_products\n(owner_id, business_id, category_id, supplier_id, brand_id, name, sku, barcode, description,\n unit, cost_price, sale_price, compare_at_price, tax_rate, currency,\n low_stock_threshold, reorder_quantity, allow_backorder, lead_time_days,\n weight, weight_unit, length, width, height, dimension_unit,\n country_of_origin, hs_code, is_active, track_inventory)\nVALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.category_id)
        .bind(params.supplier_id)
        .bind(params.brand_id)
        .bind(&params.name)
        .bind(&params.sku)
        .bind(&params.barcode)
        .bind(&params.description)
        .bind(&params.unit)
        .bind(&params.cost_price)
        .bind(&params.sale_price)
        .bind(&params.compare_at_price)
        .bind(&params.tax_rate)
        .bind(&params.currency)
        .bind(&params.low_stock_threshold)
        .bind(&params.reorder_quantity)
        .bind(params.allow_backorder)
        .bind(params.lead_time_days)
        .bind(&params.weight)
        .bind(&params.weight_unit)
        .bind(&params.length)
        .bind(&params.width)
        .bind(&params.height)
        .bind(&params.dimension_unit)
        .bind(&params.country_of_origin)
        .bind(&params.hs_code)
        .bind(params.is_active)
        .bind(params.track_inventory)
        .execute(exec)
        .await
}

/// Params for `DeletePrimaryProductImage` (Go `DeletePrimaryProductImageParams`).
pub struct DeletePrimaryProductImageParams {
    pub business_id: i64,
    pub product_id: i64,
}

/// `DeletePrimaryProductImage`: `exec`.
pub async fn delete_primary_product_image<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeletePrimaryProductImageParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM erp_product_images WHERE business_id = ? AND product_id = ? AND is_primary = TRUE")
        .bind(params.business_id)
        .bind(params.product_id)
        .execute(exec)
        .await
}

/// Params for `DeleteProduct` (Go `DeleteProductParams`).
pub struct DeleteProductParams {
    pub business_id: i64,
    pub id: i64,
}

/// `DeleteProduct`: `exec`.
pub async fn delete_product<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteProductParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM erp_products WHERE business_id = ? AND id = ?")
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `DeleteProductAttributes` (Go `DeleteProductAttributesParams`).
pub struct DeleteProductAttributesParams {
    pub business_id: i64,
    pub product_id: i64,
}

/// `DeleteProductAttributes`: `exec`.
pub async fn delete_product_attributes<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteProductAttributesParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM erp_product_attributes WHERE business_id = ? AND product_id = ?")
        .bind(params.business_id)
        .bind(params.product_id)
        .execute(exec)
        .await
}

/// Params for `GetBrandReference` (Go `GetBrandReferenceParams`).
pub struct GetBrandReferenceParams {
    pub business_id: i64,
    pub id: i64,
}

/// `GetBrandReference`: `one`.
pub async fn get_brand_reference<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetBrandReferenceParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id FROM erp_brands WHERE business_id = ? AND id = ?")
        .bind(params.business_id)
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetCategoryReference` (Go `GetCategoryReferenceParams`).
pub struct GetCategoryReferenceParams {
    pub business_id: i64,
    pub id: i64,
}

/// `GetCategoryReference`: `one`.
pub async fn get_category_reference<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetCategoryReferenceParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id FROM erp_categories WHERE business_id = ? AND id = ?")
        .bind(params.business_id)
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetProduct` (Go `GetProductParams`).
pub struct GetProductParams {
    pub business_id: i64,
    pub id: i64,
}

/// `GetProduct`: `one`.
pub async fn get_product<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetProductParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, category_id, supplier_id, brand_id, name, sku, barcode, description,\n       unit, cost_price, sale_price, compare_at_price, tax_rate, currency,\n       stock_quantity, low_stock_threshold, reorder_quantity, allow_backorder, lead_time_days,\n       weight, weight_unit, length, width, height, dimension_unit,\n       country_of_origin, hs_code, is_active, track_inventory, created_at, updated_at\nFROM erp_products\nWHERE business_id = ? AND id = ?")
        .bind(params.business_id)
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetProductByBarcode` (Go `GetProductByBarcodeParams`).
pub struct GetProductByBarcodeParams {
    pub business_id: i64,
    pub barcode: Option<String>,
}

/// `GetProductByBarcode`: `one`.
pub async fn get_product_by_barcode<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetProductByBarcodeParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, category_id, supplier_id, brand_id, name, sku, barcode, description,\n       unit, cost_price, sale_price, compare_at_price, tax_rate, currency,\n       stock_quantity, low_stock_threshold, reorder_quantity, allow_backorder, lead_time_days,\n       weight, weight_unit, length, width, height, dimension_unit,\n       country_of_origin, hs_code, is_active, track_inventory, created_at, updated_at\nFROM erp_products\nWHERE business_id = ? AND barcode = ?\nLIMIT 1")
        .bind(params.business_id)
        .bind(&params.barcode)
        .fetch_optional(exec)
        .await
}

/// Params for `GetProductForUpdate` (Go `GetProductForUpdateParams`).
pub struct GetProductForUpdateParams {
    pub business_id: i64,
    pub id: i64,
}

/// `GetProductForUpdate`: `one`.
pub async fn get_product_for_update<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetProductForUpdateParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, category_id, supplier_id, brand_id, name, sku, barcode, description,\n       unit, cost_price, sale_price, compare_at_price, tax_rate, currency,\n       stock_quantity, low_stock_threshold, reorder_quantity, allow_backorder, lead_time_days,\n       weight, weight_unit, length, width, height, dimension_unit,\n       country_of_origin, hs_code, is_active, track_inventory, created_at, updated_at\nFROM erp_products\nWHERE business_id = ? AND id = ? FOR UPDATE")
        .bind(params.business_id)
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetSalesProduct` (Go `GetSalesProductParams`).
pub struct GetSalesProductParams {
    pub business_id: i64,
    pub id: i64,
}

/// `GetSalesProduct`: `one`.
pub async fn get_sales_product<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetSalesProductParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT name, COALESCE(sku, '') AS sku, sale_price, COALESCE(tax_rate, 0) AS tax_rate\nFROM erp_products WHERE business_id = ? AND id = ? AND is_active = TRUE")
        .bind(params.business_id)
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetSupplierReference` (Go `GetSupplierReferenceParams`).
pub struct GetSupplierReferenceParams {
    pub business_id: i64,
    pub id: i64,
}

/// `GetSupplierReference`: `one`.
pub async fn get_supplier_reference<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetSupplierReferenceParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id FROM erp_suppliers WHERE business_id = ? AND id = ?")
        .bind(params.business_id)
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `InsertPrimaryProductImage` (Go `InsertPrimaryProductImageParams`).
pub struct InsertPrimaryProductImageParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub product_id: i64,
    pub url: String,
    pub alt: Option<String>,
}

/// `InsertPrimaryProductImage`: `exec`.
pub async fn insert_primary_product_image<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsertPrimaryProductImageParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO erp_product_images (owner_id, business_id, product_id, url, alt, position, is_primary)\nVALUES (?, ?, ?, ?, ?, 0, TRUE)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.product_id)
        .bind(&params.url)
        .bind(&params.alt)
        .execute(exec)
        .await
}

/// Params for `InsertProductAttribute` (Go `InsertProductAttributeParams`).
pub struct InsertProductAttributeParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub product_id: i64,
    pub name: String,
    pub value: String,
    pub position: i32,
}

/// `InsertProductAttribute`: `exec`.
pub async fn insert_product_attribute<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsertProductAttributeParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO erp_product_attributes (owner_id, business_id, product_id, name, value, position)\nVALUES (?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.product_id)
        .bind(&params.name)
        .bind(&params.value)
        .bind(params.position)
        .execute(exec)
        .await
}

/// Params for `InsightsActiveProducts` (Go `InsightsActiveProductsParams`).
pub struct InsightsActiveProductsParams {
    pub business_id: i64,
}

/// `InsightsActiveProducts`: `many`.
pub async fn insights_active_products<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsightsActiveProductsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, name, stock_quantity\nFROM erp_products\nWHERE business_id = ?\n  AND is_active = TRUE\nORDER BY name ASC, id ASC")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `InsightsAvgMarginPct` (Go `InsightsAvgMarginPctParams`).
pub struct InsightsAvgMarginPctParams {
    pub business_id: i64,
}

/// `InsightsAvgMarginPct`: `one`.
pub async fn insights_avg_margin_pct<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsightsAvgMarginPctParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COALESCE(AVG((sale_price - cost_price) / sale_price * 100), 0) AS margin_pct\nFROM erp_products\nWHERE business_id = ?\n  AND is_active = TRUE\n  AND sale_price > 0")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `InsightsCountActiveProducts` (Go `InsightsCountActiveProductsParams`).
pub struct InsightsCountActiveProductsParams {
    pub business_id: i64,
}

/// `InsightsCountActiveProducts`: `one`.
pub async fn insights_count_active_products<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsightsCountActiveProductsParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query(
        "SELECT COUNT(*) AS count FROM erp_products WHERE business_id = ? AND is_active = TRUE",
    )
    .bind(params.business_id)
    .fetch_optional(exec)
    .await
}

/// Params for `InsightsCountLowStock` (Go `InsightsCountLowStockParams`).
pub struct InsightsCountLowStockParams {
    pub business_id: i64,
}

/// `InsightsCountLowStock`: `one`.
pub async fn insights_count_low_stock<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsightsCountLowStockParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) AS count\nFROM erp_products\nWHERE business_id = ? AND is_active = TRUE AND track_inventory = TRUE\n  AND low_stock_threshold IS NOT NULL\n  AND stock_quantity <= low_stock_threshold")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `InsightsCountOutOfStock` (Go `InsightsCountOutOfStockParams`).
pub struct InsightsCountOutOfStockParams {
    pub business_id: i64,
}

/// `InsightsCountOutOfStock`: `one`.
pub async fn insights_count_out_of_stock<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsightsCountOutOfStockParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) AS count\nFROM erp_products\nWHERE business_id = ?\n  AND track_inventory = TRUE\n  AND stock_quantity <= 0")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `InsightsCountProducts` (Go `InsightsCountProductsParams`).
pub struct InsightsCountProductsParams {
    pub business_id: i64,
}

/// `InsightsCountProducts`: `one`.
pub async fn insights_count_products<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsightsCountProductsParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) AS count FROM erp_products WHERE business_id = ?")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `InsightsERPLastChange` (Go `InsightsERPLastChangeParams`).
pub struct InsightsERPLastChangeParams {
    pub business_id: i64,
}

/// `InsightsERPLastChange`: `one`.
pub async fn insights_erp_last_change<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsightsERPLastChangeParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT CAST(COALESCE(MAX(updated_at), '1000-01-01') AS DATETIME) AS last_change\nFROM erp_products\nWHERE business_id = ?")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `InsightsMarginDistribution` (Go `InsightsMarginDistributionParams`).
pub struct InsightsMarginDistributionParams {
    pub business_id: i64,
}

/// `InsightsMarginDistribution`: `one`.
pub async fn insights_margin_distribution<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsightsMarginDistributionParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT\n  CAST(COALESCE(SUM(margin_pct >= -1000 AND margin_pct < 0), 0) AS SIGNED) AS below_zero_count,\n  CAST(COALESCE(SUM(margin_pct >= 0 AND margin_pct < 10), 0) AS SIGNED) AS zero_to_ten_count,\n  CAST(COALESCE(SUM(margin_pct >= 10 AND margin_pct < 25), 0) AS SIGNED) AS ten_to_twenty_five_count,\n  CAST(COALESCE(SUM(margin_pct >= 25 AND margin_pct < 50), 0) AS SIGNED) AS twenty_five_to_fifty_count,\n  CAST(COALESCE(SUM(margin_pct >= 50 AND margin_pct < 75), 0) AS SIGNED) AS fifty_to_seventy_five_count,\n  CAST(COALESCE(SUM(margin_pct >= 75 AND margin_pct < 1000), 0) AS SIGNED) AS seventy_five_plus_count\nFROM (\n  SELECT (sale_price - cost_price) / sale_price * 100 AS margin_pct\n  FROM erp_products\n  WHERE business_id = ?\n    AND is_active = TRUE\n    AND sale_price > 0\n) AS active_product_margins")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `InsightsProductFacts` (Go `InsightsProductFactsParams`).
pub struct InsightsProductFactsParams {
    pub business_id: i64,
    pub ids: Vec<i64>,
}

/// `InsightsProductFacts`: `many`.
pub async fn insights_product_facts<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsightsProductFactsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    let mut sql_template: String = "SELECT p.id,\n       p.name,\n       COALESCE(c.id, 0) AS category_id,\n       COALESCE(c.name, '') AS category_name,\n       CAST(p.cost_price * 100 AS SIGNED) AS cost_minor_units\nFROM erp_products p\nLEFT JOIN erp_categories c ON c.id = p.category_id AND c.business_id = p.business_id\nWHERE p.business_id = ?\n  AND p.id IN ({SLICE:ids})".to_string();

    if params.ids.is_empty() {
        sql_template = sql_template.replace("{SLICE:ids}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:ids}",
            &std::iter::repeat_n("?", params.ids.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    let query = sqlx::query(&sql_template);
    let mut query = query.bind(params.business_id);
    for value in &params.ids {
        query = query.bind(value);
    }
    query.execute(exec).await
}

/// Params for `ListProductAttributes` (Go `ListProductAttributesParams`).
pub struct ListProductAttributesParams {
    pub business_id: i64,
    pub product_id: i64,
}

/// `ListProductAttributes`: `many`.
pub async fn list_product_attributes<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListProductAttributesParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, name, value, position FROM erp_product_attributes\nWHERE business_id = ? AND product_id = ?\nORDER BY position ASC, id ASC")
        .bind(params.business_id)
        .bind(params.product_id)
        .execute(exec)
        .await
}

/// Params for `ListRecentPurchaseOrders` (Go `ListRecentPurchaseOrdersParams`).
pub struct ListRecentPurchaseOrdersParams {
    pub business_id: i64,
    pub limit: i32,
}

/// `ListRecentPurchaseOrders`: `many`.
pub async fn list_recent_purchase_orders<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListRecentPurchaseOrdersParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT po.id, po.po_number, COALESCE(s.name, 'Unknown') AS supplier_name,\n       po.status, po.total, po.created_at\nFROM erp_purchase_orders po\nLEFT JOIN erp_suppliers s ON s.id = po.supplier_id\nWHERE po.business_id = ?\nORDER BY po.created_at DESC\nLIMIT ?")
        .bind(params.business_id)
        .bind(params.limit)
        .execute(exec)
        .await
}

/// Params for `ListReorderSuggestions` (Go `ListReorderSuggestionsParams`).
pub struct ListReorderSuggestionsParams {
    pub business_id: i64,
}

/// `ListReorderSuggestions`: `many`.
pub async fn list_reorder_suggestions<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListReorderSuggestionsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, name, COALESCE(sku, '') AS sku,\n       stock_quantity, COALESCE(low_stock_threshold, 0) AS low_stock_threshold,\n       COALESCE(reorder_quantity, 0) AS reorder_quantity\nFROM erp_products\nWHERE business_id = ? AND is_active = TRUE AND track_inventory = TRUE\n  AND low_stock_threshold IS NOT NULL\n  AND stock_quantity <= low_stock_threshold\nORDER BY (stock_quantity - low_stock_threshold) ASC, id ASC\nLIMIT 101")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `ListTopCategories` (Go `ListTopCategoriesParams`).
pub struct ListTopCategoriesParams {
    pub business_id: i64,
    pub limit: i32,
}

/// `ListTopCategories`: `many`.
pub async fn list_top_categories<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListTopCategoriesParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT COALESCE(c.name, 'Uncategorized') AS name, COUNT(p.id) AS product_count\nFROM erp_products p\nLEFT JOIN erp_categories c ON c.id = p.category_id\nWHERE p.business_id = ? AND p.is_active = TRUE\nGROUP BY c.id, c.name\nORDER BY product_count DESC\nLIMIT ?")
        .bind(params.business_id)
        .bind(params.limit)
        .execute(exec)
        .await
}

/// Params for `ListTopLowStockItems` (Go `ListTopLowStockItemsParams`).
pub struct ListTopLowStockItemsParams {
    pub business_id: i64,
    pub limit: i32,
}

/// `ListTopLowStockItems`: `many`.
pub async fn list_top_low_stock_items<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListTopLowStockItemsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT p.id, p.name, COALESCE(p.sku, '') AS sku,\n       p.stock_quantity, COALESCE(p.low_stock_threshold, 0) AS low_stock_threshold\nFROM erp_products p\nWHERE p.business_id = ? AND p.is_active = TRUE AND p.track_inventory = TRUE\n  AND p.low_stock_threshold IS NOT NULL\n  AND p.stock_quantity <= p.low_stock_threshold\nORDER BY (p.stock_quantity - p.low_stock_threshold) ASC, p.stock_quantity ASC\nLIMIT ?")
        .bind(params.business_id)
        .bind(params.limit)
        .execute(exec)
        .await
}

/// Params for `SumInventoryValue` (Go `SumInventoryValueParams`).
pub struct SumInventoryValueParams {
    pub business_id: i64,
}

/// `SumInventoryValue`: `one`.
pub async fn sum_inventory_value<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &SumInventoryValueParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COALESCE(SUM(stock_quantity * cost_price), 0) AS inventory_value\nFROM erp_products\nWHERE business_id = ? AND is_active = TRUE AND track_inventory = TRUE")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `UpdateProduct` (Go `UpdateProductParams`).
pub struct UpdateProductParams {
    pub category_id: Option<i64>,
    pub supplier_id: Option<i64>,
    pub brand_id: Option<i64>,
    pub name: String,
    pub sku: Option<String>,
    pub barcode: Option<String>,
    pub description: Option<String>,
    pub unit: Option<String>,
    pub cost_price: String,
    pub sale_price: String,
    pub compare_at_price: String,
    pub tax_rate: String,
    pub currency: String,
    pub low_stock_threshold: Option<String>,
    pub reorder_quantity: Option<String>,
    pub allow_backorder: bool,
    pub lead_time_days: i32,
    pub weight: String,
    pub weight_unit: String,
    pub length: String,
    pub width: String,
    pub height: String,
    pub dimension_unit: String,
    pub country_of_origin: Option<String>,
    pub hs_code: Option<String>,
    pub is_active: Option<bool>,
    pub track_inventory: Option<bool>,
    pub business_id: i64,
    pub id: i64,
}

/// `UpdateProduct`: `exec`.
pub async fn update_product<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateProductParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE erp_products\nSET category_id = ?, supplier_id = ?, brand_id = ?, name = ?, sku = ?, barcode = ?,\n    description = ?, unit = ?, cost_price = ?, sale_price = ?, compare_at_price = ?,\n    tax_rate = ?, currency = ?, low_stock_threshold = ?,\n    reorder_quantity = ?, allow_backorder = ?, lead_time_days = ?, weight = ?,\n    weight_unit = ?, length = ?, width = ?, height = ?, dimension_unit = ?,\n    country_of_origin = ?, hs_code = ?, is_active = ?, track_inventory = ?\nWHERE business_id = ? AND id = ?")
        .bind(params.category_id)
        .bind(params.supplier_id)
        .bind(params.brand_id)
        .bind(&params.name)
        .bind(&params.sku)
        .bind(&params.barcode)
        .bind(&params.description)
        .bind(&params.unit)
        .bind(&params.cost_price)
        .bind(&params.sale_price)
        .bind(&params.compare_at_price)
        .bind(&params.tax_rate)
        .bind(&params.currency)
        .bind(&params.low_stock_threshold)
        .bind(&params.reorder_quantity)
        .bind(params.allow_backorder)
        .bind(params.lead_time_days)
        .bind(&params.weight)
        .bind(&params.weight_unit)
        .bind(&params.length)
        .bind(&params.width)
        .bind(&params.height)
        .bind(&params.dimension_unit)
        .bind(&params.country_of_origin)
        .bind(&params.hs_code)
        .bind(params.is_active)
        .bind(params.track_inventory)
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}
