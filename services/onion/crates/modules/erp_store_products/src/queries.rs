//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CreateStoreProduct` (Go `CreateStoreProductParams`).
pub struct CreateStoreProductParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub product_id: Option<i64>,
    pub title: String,
    pub slug: Option<String>,
    pub description: Option<String>,
    pub images: String,
    pub price: String,
    pub compare_at_price: Option<String>,
    pub is_published: Option<bool>,
    pub sort_order: Option<i32>,
}

/// `CreateStoreProduct`: `execresult`.
pub async fn create_store_product<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateStoreProductParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO store_products (owner_id, business_id, product_id, title, slug, description, images, price, compare_at_price, is_published, sort_order)\nVALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.product_id)
        .bind(&params.title)
        .bind(&params.slug)
        .bind(&params.description)
        .bind(&params.images)
        .bind(&params.price)
        .bind(&params.compare_at_price)
        .bind(params.is_published)
        .bind(params.sort_order)
        .execute(exec)
        .await
}

/// Params for `DeleteCatalogProductImages` (Go `DeleteCatalogProductImagesParams`).
pub struct DeleteCatalogProductImagesParams {
    pub business_id: i64,
    pub product_id: i64,
}

/// `DeleteCatalogProductImages`: `exec`.
pub async fn delete_catalog_product_images<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteCatalogProductImagesParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM erp_product_images\nWHERE business_id = ? AND product_id = ?")
        .bind(params.business_id)
        .bind(params.product_id)
        .execute(exec)
        .await
}

/// Params for `DeleteStoreProduct` (Go `DeleteStoreProductParams`).
pub struct DeleteStoreProductParams {
    pub business_id: i64,
    pub id: i64,
}

/// `DeleteStoreProduct`: `exec`.
pub async fn delete_store_product<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteStoreProductParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM store_products WHERE business_id = ? AND id = ?")
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `GetCatalogProductByID` (Go `GetCatalogProductByIDParams`).
pub struct GetCatalogProductByIDParams {
    pub business_id: i64,
    pub id: i64,
}

/// `GetCatalogProductByID`: `one`.
pub async fn get_catalog_product_by_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetCatalogProductByIDParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query(
        "SELECT id\nFROM erp_products\nWHERE business_id = ? AND id = ?\nORDER BY id ASC\nLIMIT 1",
    )
    .bind(params.business_id)
    .bind(params.id)
    .fetch_optional(exec)
    .await
}

/// Params for `GetStoreProductLinkedID` (Go `GetStoreProductLinkedIDParams`).
pub struct GetStoreProductLinkedIDParams {
    pub business_id: i64,
    pub id: i64,
}

/// `GetStoreProductLinkedID`: `one`.
pub async fn get_store_product_linked_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetStoreProductLinkedIDParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT product_id\nFROM store_products\nWHERE business_id = ? AND id = ?\nORDER BY id ASC\nLIMIT 1")
        .bind(params.business_id)
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `InsertCatalogProduct` (Go `InsertCatalogProductParams`).
pub struct InsertCatalogProductParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub name: String,
    pub description: Option<String>,
    pub sale_price: String,
    pub compare_at_price: String,
}

/// `InsertCatalogProduct`: `execresult`.
pub async fn insert_catalog_product<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsertCatalogProductParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO erp_products (\n    owner_id,\n    business_id,\n    name,\n    description,\n    sale_price,\n    compare_at_price,\n    is_active,\n    track_inventory\n) VALUES (?, ?, ?, ?, ?, ?, TRUE, TRUE)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(&params.name)
        .bind(&params.description)
        .bind(&params.sale_price)
        .bind(&params.compare_at_price)
        .execute(exec)
        .await
}

/// Params for `InsertCatalogProductImage` (Go `InsertCatalogProductImageParams`).
pub struct InsertCatalogProductImageParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub product_id: i64,
    pub url: String,
    pub alt: Option<String>,
    pub position: i32,
    pub is_primary: bool,
}

/// `InsertCatalogProductImage`: `exec`.
pub async fn insert_catalog_product_image<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsertCatalogProductImageParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO erp_product_images (\n    owner_id,\n    business_id,\n    product_id,\n    url,\n    alt,\n    position,\n    is_primary\n) VALUES (?, ?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.product_id)
        .bind(&params.url)
        .bind(&params.alt)
        .bind(params.position)
        .bind(params.is_primary)
        .execute(exec)
        .await
}

/// Params for `ListStoreProductsWithCatalog` (Go `ListStoreProductsWithCatalogParams`).
pub struct ListStoreProductsWithCatalogParams {
    pub business_id: i64,
}

/// `ListStoreProductsWithCatalog`: `many`.
pub async fn list_store_products_with_catalog<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListStoreProductsWithCatalogParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("WITH store_product_page AS (\n    SELECT\n        id,\n        owner_id,\n        business_id,\n        product_id,\n        title,\n        slug,\n        description,\n        images,\n        price,\n        compare_at_price,\n        is_published,\n        sort_order,\n        created_at,\n        updated_at\n    FROM store_products store_product\n    WHERE store_product.business_id = ?\n    ORDER BY store_product.sort_order ASC, store_product.id ASC\n    LIMIT 200\n),\ncatalog_product_page AS (\n    SELECT DISTINCT page.business_id, page.product_id\n    FROM store_product_page page\n    WHERE page.product_id IS NOT NULL\n),\ncatalog_images AS (\n    SELECT images.business_id, images.product_id, JSON_ARRAYAGG(images.url) AS images\n    FROM erp_product_images images\n    INNER JOIN catalog_product_page page\n      ON page.business_id = images.business_id AND page.product_id = images.product_id\n    GROUP BY images.business_id, images.product_id\n)\nSELECT\n    sp.id,\n    sp.owner_id,\n    sp.business_id,\n    sp.product_id,\n    COALESCE(ep.name, sp.title) AS title,\n    sp.slug,\n    COALESCE(ep.description, sp.description) AS description,\n    COALESCE(epi.images, sp.images, JSON_ARRAY()) AS images,\n    COALESCE(ep.sale_price, sp.price) AS price,\n    COALESCE(ep.compare_at_price, sp.compare_at_price) AS compare_at_price,\n    sp.is_published,\n    sp.sort_order,\n    sp.created_at,\n    sp.updated_at\nFROM store_product_page sp\nLEFT JOIN erp_products ep\n  ON ep.business_id = sp.business_id AND ep.id = sp.product_id\nLEFT JOIN catalog_images epi\n  ON epi.business_id = sp.business_id AND epi.product_id = sp.product_id\nORDER BY sp.sort_order ASC, sp.id ASC")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpdateCatalogProduct` (Go `UpdateCatalogProductParams`).
pub struct UpdateCatalogProductParams {
    pub name: String,
    pub description: Option<String>,
    pub sale_price: String,
    pub compare_at_price: String,
    pub business_id: i64,
    pub id: i64,
}

/// `UpdateCatalogProduct`: `exec`.
pub async fn update_catalog_product<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateCatalogProductParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE erp_products\nSET\n    name = ?,\n    description = ?,\n    sale_price = ?,\n    compare_at_price = ?,\n    is_active = TRUE\nWHERE business_id = ? AND id = ?")
        .bind(&params.name)
        .bind(&params.description)
        .bind(&params.sale_price)
        .bind(&params.compare_at_price)
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `UpdateStoreProduct` (Go `UpdateStoreProductParams`).
pub struct UpdateStoreProductParams {
    pub product_id: Option<i64>,
    pub title: String,
    pub slug: Option<String>,
    pub description: Option<String>,
    pub images: String,
    pub price: String,
    pub compare_at_price: Option<String>,
    pub is_published: Option<bool>,
    pub sort_order: Option<i32>,
    pub business_id: i64,
    pub id: i64,
}

/// `UpdateStoreProduct`: `execrows`.
pub async fn update_store_product<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateStoreProductParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE store_products SET product_id = ?, title = ?, slug = ?, description = ?, images = ?, price = ?, compare_at_price = ?, is_published = ?, sort_order = ?\nWHERE business_id = ? AND id = ?")
        .bind(params.product_id)
        .bind(&params.title)
        .bind(&params.slug)
        .bind(&params.description)
        .bind(&params.images)
        .bind(&params.price)
        .bind(&params.compare_at_price)
        .bind(params.is_published)
        .bind(params.sort_order)
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}
