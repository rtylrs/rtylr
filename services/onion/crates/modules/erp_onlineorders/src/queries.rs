//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CountStoreOrders` (Go `CountStoreOrdersParams`).
pub struct CountStoreOrdersParams {
    pub business_id: i64,
}

/// `CountStoreOrders`: `one`.
pub async fn count_store_orders<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountStoreOrdersParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM store_orders WHERE business_id = ?")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CreateStoreOrder` (Go `CreateStoreOrderParams`).
pub struct CreateStoreOrderParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub customer_email: Option<String>,
    pub customer_name: Option<String>,
    pub customer_phone: Option<String>,
    pub status: String,
    pub subtotal: String,
    pub tax: String,
    pub shipping_cost: String,
    pub total: String,
    pub shipping_address: String,
    pub notes: Option<String>,
}

/// `CreateStoreOrder`: `execresult`.
pub async fn create_store_order<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateStoreOrderParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO store_orders (owner_id, business_id, customer_email, customer_name, customer_phone, status, subtotal, tax, shipping_cost, total, shipping_address, notes)\nVALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(&params.customer_email)
        .bind(&params.customer_name)
        .bind(&params.customer_phone)
        .bind(&params.status)
        .bind(&params.subtotal)
        .bind(&params.tax)
        .bind(&params.shipping_cost)
        .bind(&params.total)
        .bind(&params.shipping_address)
        .bind(&params.notes)
        .execute(exec)
        .await
}

/// Params for `CreateStoreOrderItem` (Go `CreateStoreOrderItemParams`).
pub struct CreateStoreOrderItemParams {
    pub order_id: i64,
    pub product_id: Option<i64>,
    pub name: String,
    pub quantity: i32,
    pub price: String,
    pub total: String,
}

/// `CreateStoreOrderItem`: `execresult`.
pub async fn create_store_order_item<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateStoreOrderItemParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO store_order_items (order_id, product_id, name, quantity, price, total)\nVALUES (?, ?, ?, ?, ?, ?)")
        .bind(params.order_id)
        .bind(params.product_id)
        .bind(&params.name)
        .bind(params.quantity)
        .bind(&params.price)
        .bind(&params.total)
        .execute(exec)
        .await
}

/// Params for `GetStoreOrder` (Go `GetStoreOrderParams`).
pub struct GetStoreOrderParams {
    pub business_id: i64,
    pub id: i64,
}

/// `GetStoreOrder`: `one`.
pub async fn get_store_order<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetStoreOrderParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, customer_email, customer_name, customer_phone, status, subtotal, tax, shipping_cost, total, shipping_address, notes, created_at, updated_at\nFROM store_orders WHERE business_id = ? AND id = ?")
        .bind(params.business_id)
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetStoreOrderStatus` (Go `GetStoreOrderStatusParams`).
pub struct GetStoreOrderStatusParams {
    pub business_id: i64,
    pub id: i64,
}

/// `GetStoreOrderStatus`: `one`.
pub async fn get_store_order_status<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetStoreOrderStatusParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query(
        "SELECT status, total FROM store_orders WHERE business_id = ? AND id = ? FOR UPDATE",
    )
    .bind(params.business_id)
    .bind(params.id)
    .fetch_optional(exec)
    .await
}

/// Params for `GetStoreProductPricing` (Go `GetStoreProductPricingParams`).
pub struct GetStoreProductPricingParams {
    pub business_id: i64,
    pub id: i64,
}

/// `GetStoreProductPricing`: `one`.
pub async fn get_store_product_pricing<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetStoreProductPricingParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT name, sale_price, tax_rate, track_inventory\nFROM erp_products\nWHERE business_id = ? AND id = ? AND is_active = TRUE")
        .bind(params.business_id)
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `ListStoreOrderItems` (Go `ListStoreOrderItemsParams`).
pub struct ListStoreOrderItemsParams {
    pub business_id: i64,
    pub order_id: i64,
}

/// `ListStoreOrderItems`: `many`.
pub async fn list_store_order_items<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListStoreOrderItemsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT store_order_items.id, store_order_items.order_id, store_order_items.product_id, store_order_items.name, store_order_items.quantity, store_order_items.price, store_order_items.total, store_order_items.created_at\nFROM store_order_items\nINNER JOIN store_orders ON store_order_items.order_id = store_orders.id\nWHERE store_orders.business_id = ? AND store_order_items.order_id = ?\nORDER BY store_order_items.id ASC\nLIMIT 500")
        .bind(params.business_id)
        .bind(params.order_id)
        .execute(exec)
        .await
}

/// Params for `ListStoreOrderItemsForUpdate` (Go `ListStoreOrderItemsForUpdateParams`).
pub struct ListStoreOrderItemsForUpdateParams {
    pub business_id: i64,
    pub order_id: i64,
}

/// `ListStoreOrderItemsForUpdate`: `many`.
pub async fn list_store_order_items_for_update<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListStoreOrderItemsForUpdateParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT store_order_items.id, store_order_items.order_id, store_order_items.product_id, store_order_items.name, store_order_items.quantity, store_order_items.price, store_order_items.total, store_order_items.created_at\nFROM store_order_items\nINNER JOIN store_orders ON store_order_items.order_id = store_orders.id\nWHERE store_orders.business_id = ? AND store_order_items.order_id = ?\nORDER BY store_order_items.id ASC\nLIMIT 500\nFOR UPDATE")
        .bind(params.business_id)
        .bind(params.order_id)
        .execute(exec)
        .await
}

/// Params for `ListStoreOrders` (Go `ListStoreOrdersParams`).
pub struct ListStoreOrdersParams {
    pub business_id: i64,
}

/// `ListStoreOrders`: `many`.
pub async fn list_store_orders<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListStoreOrdersParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, customer_email, customer_name, customer_phone, status, subtotal, tax, shipping_cost, total, shipping_address, notes, created_at, updated_at\nFROM store_orders WHERE business_id = ? ORDER BY created_at DESC, id DESC LIMIT 500")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `ListStoreOrdersPaginated` (Go `ListStoreOrdersPaginatedParams`).
pub struct ListStoreOrdersPaginatedParams {
    pub business_id: i64,
    pub limit: i32,
    pub offset: i32,
}

/// `ListStoreOrdersPaginated`: `many`.
pub async fn list_store_orders_paginated<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListStoreOrdersPaginatedParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, customer_email, customer_name, customer_phone, status, subtotal, tax, shipping_cost, total, shipping_address, notes, created_at, updated_at\nFROM store_orders WHERE business_id = ? ORDER BY created_at DESC, id DESC\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `UpdateStoreOrder` (Go `UpdateStoreOrderParams`).
pub struct UpdateStoreOrderParams {
    pub customer_email: Option<String>,
    pub customer_name: Option<String>,
    pub customer_phone: Option<String>,
    pub status: String,
    pub shipping_address: String,
    pub notes: Option<String>,
    pub business_id: i64,
    pub id: i64,
}

/// `UpdateStoreOrder`: `exec`.
pub async fn update_store_order<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateStoreOrderParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE store_orders SET customer_email = ?, customer_name = ?, customer_phone = ?, status = ?, shipping_address = ?, notes = ?\nWHERE business_id = ? AND id = ?")
        .bind(&params.customer_email)
        .bind(&params.customer_name)
        .bind(&params.customer_phone)
        .bind(&params.status)
        .bind(&params.shipping_address)
        .bind(&params.notes)
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}
