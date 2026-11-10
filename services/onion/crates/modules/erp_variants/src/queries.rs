//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CountVariantsByOwner` (Go `CountVariantsByOwnerParams`).
pub struct CountVariantsByOwnerParams {
    pub business_id: i64,
}

/// `CountVariantsByOwner`: `one`.
pub async fn count_variants_by_owner<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountVariantsByOwnerParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM erp_product_variants WHERE business_id = ?")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CountVariantsByProduct` (Go `CountVariantsByProductParams`).
pub struct CountVariantsByProductParams {
    pub business_id: i64,
    pub product_id: i64,
}

/// `CountVariantsByProduct`: `one`.
pub async fn count_variants_by_product<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountVariantsByProductParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query(
        "SELECT COUNT(*) FROM erp_product_variants WHERE business_id = ? AND product_id = ?",
    )
    .bind(params.business_id)
    .bind(params.product_id)
    .fetch_optional(exec)
    .await
}

/// Params for `CreateVariant` (Go `CreateVariantParams`).
pub struct CreateVariantParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub product_id: i64,
    pub name: String,
    pub sku: Option<String>,
    pub barcode: Option<String>,
    pub cost_price: String,
    pub sale_price: String,
    pub stock_quantity: String,
    pub attributes: String,
    pub is_active: Option<bool>,
}

/// `CreateVariant`: `execresult`.
pub async fn create_variant<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateVariantParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO erp_product_variants\n(owner_id, business_id, product_id, name, sku, barcode, cost_price, sale_price, stock_quantity, attributes, is_active)\nVALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.product_id)
        .bind(&params.name)
        .bind(&params.sku)
        .bind(&params.barcode)
        .bind(&params.cost_price)
        .bind(&params.sale_price)
        .bind(&params.stock_quantity)
        .bind(&params.attributes)
        .bind(params.is_active)
        .execute(exec)
        .await
}

/// Params for `DeleteVariant` (Go `DeleteVariantParams`).
pub struct DeleteVariantParams {
    pub business_id: i64,
    pub id: i64,
}

/// `DeleteVariant`: `execrows`.
pub async fn delete_variant<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteVariantParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM erp_product_variants WHERE business_id = ? AND id = ?")
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}

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

/// Params for `ListVariantsByOwner` (Go `ListVariantsByOwnerParams`).
pub struct ListVariantsByOwnerParams {
    pub business_id: i64,
    pub limit: i32,
    pub offset: i32,
}

/// `ListVariantsByOwner`: `many`.
pub async fn list_variants_by_owner<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListVariantsByOwnerParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, product_id, name, sku, barcode,\n       cost_price, sale_price, stock_quantity, COALESCE(attributes, '[]') AS attributes, is_active,\n       created_at, updated_at\nFROM erp_product_variants\nWHERE business_id = ?\nORDER BY name ASC, id ASC\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `ListVariantsByProduct` (Go `ListVariantsByProductParams`).
pub struct ListVariantsByProductParams {
    pub business_id: i64,
    pub product_id: i64,
    pub limit: i32,
    pub offset: i32,
}

/// `ListVariantsByProduct`: `many`.
pub async fn list_variants_by_product<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListVariantsByProductParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, product_id, name, sku, barcode,\n       cost_price, sale_price, stock_quantity, COALESCE(attributes, '[]') AS attributes, is_active,\n       created_at, updated_at\nFROM erp_product_variants\nWHERE business_id = ? AND product_id = ?\nORDER BY name ASC, id ASC\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(params.product_id)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `UpdateVariant` (Go `UpdateVariantParams`).
pub struct UpdateVariantParams {
    pub product_id: i64,
    pub name: String,
    pub sku: Option<String>,
    pub barcode: Option<String>,
    pub cost_price: String,
    pub sale_price: String,
    pub stock_quantity: String,
    pub attributes: String,
    pub is_active: Option<bool>,
    pub business_id: i64,
    pub id: i64,
}

/// `UpdateVariant`: `execrows`.
pub async fn update_variant<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateVariantParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE erp_product_variants\nSET product_id = ?, name = ?, sku = ?, barcode = ?,\n    cost_price = ?, sale_price = ?, stock_quantity = ?, attributes = ?, is_active = ?\nWHERE business_id = ? AND id = ?")
        .bind(params.product_id)
        .bind(&params.name)
        .bind(&params.sku)
        .bind(&params.barcode)
        .bind(&params.cost_price)
        .bind(&params.sale_price)
        .bind(&params.stock_quantity)
        .bind(&params.attributes)
        .bind(params.is_active)
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}
