//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `ApprovePurchaseOrder` (Go `ApprovePurchaseOrderParams`).
pub struct ApprovePurchaseOrderParams {
    pub business_id: i64,
    pub id: i64,
}

/// `ApprovePurchaseOrder`: `execresult`.
pub async fn approve_purchase_order<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ApprovePurchaseOrderParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE erp_purchase_orders\nSET status = 'approved'\nWHERE business_id = ? AND id = ? AND status IN ('draft', 'sent', 'partial')")
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

/// Params for `GetPurchaseOrderByID` (Go `GetPurchaseOrderByIDParams`).
pub struct GetPurchaseOrderByIDParams {
    pub id: i64,
    pub business_id: i64,
}

/// `GetPurchaseOrderByID`: `one`.
pub async fn get_purchase_order_by_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetPurchaseOrderByIDParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, supplier_id, po_number, status, subtotal, tax_amount, total,\n       expected_delivery, received_at, notes, created_by, created_at, updated_at, destination_location_id\nFROM erp_purchase_orders\nWHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetPurchaseOrderForUpdate` (Go `GetPurchaseOrderForUpdateParams`).
pub struct GetPurchaseOrderForUpdateParams {
    pub business_id: i64,
    pub id: i64,
}

/// `GetPurchaseOrderForUpdate`: `one`.
pub async fn get_purchase_order_for_update<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetPurchaseOrderForUpdateParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query(
        "SELECT status FROM erp_purchase_orders WHERE business_id = ? AND id = ? FOR UPDATE",
    )
    .bind(params.business_id)
    .bind(params.id)
    .fetch_optional(exec)
    .await
}

/// Params for `GetPurchaseOrderItems` (Go `GetPurchaseOrderItemsParams`).
pub struct GetPurchaseOrderItemsParams {
    pub purchase_order_id: i64,
}

/// `GetPurchaseOrderItems`: `many`.
pub async fn get_purchase_order_items<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetPurchaseOrderItemsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, purchase_order_id, product_id, quantity_ordered, quantity_received, unit_cost, total_cost\nFROM erp_purchase_order_items\nWHERE purchase_order_id = ?")
        .bind(params.purchase_order_id)
        .execute(exec)
        .await
}

/// Params for `GetPurchaseOrderStatus` (Go `GetPurchaseOrderStatusParams`).
pub struct GetPurchaseOrderStatusParams {
    pub business_id: i64,
    pub id: i64,
}

/// `GetPurchaseOrderStatus`: `one`.
pub async fn get_purchase_order_status<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetPurchaseOrderStatusParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT status FROM erp_purchase_orders WHERE business_id = ? AND id = ?")
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

/// Params for `InsertPurchaseOrder` (Go `InsertPurchaseOrderParams`).
pub struct InsertPurchaseOrderParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub supplier_id: i64,
    pub po_number: String,
    pub subtotal: Option<String>,
    pub tax_amount: Option<String>,
    pub total: Option<String>,
    pub expected_delivery: Option<chrono::DateTime<chrono::Utc>>,
    pub notes: Option<String>,
    pub created_by: Option<i64>,
}

/// `InsertPurchaseOrder`: `execresult`.
pub async fn insert_purchase_order<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsertPurchaseOrderParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO erp_purchase_orders\n(owner_id, business_id, supplier_id, po_number, status, subtotal, tax_amount, total,\n expected_delivery, notes, created_by)\nVALUES (?, ?, ?, ?, 'draft', ?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.supplier_id)
        .bind(&params.po_number)
        .bind(&params.subtotal)
        .bind(&params.tax_amount)
        .bind(&params.total)
        .bind(params.expected_delivery)
        .bind(&params.notes)
        .bind(params.created_by)
        .execute(exec)
        .await
}

/// Params for `InsertPurchaseOrderItem` (Go `InsertPurchaseOrderItemParams`).
pub struct InsertPurchaseOrderItemParams {
    pub purchase_order_id: i64,
    pub product_id: i64,
    pub quantity_ordered: String,
    pub unit_cost: String,
    pub total_cost: String,
}

/// `InsertPurchaseOrderItem`: `execresult`.
pub async fn insert_purchase_order_item<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsertPurchaseOrderItemParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO erp_purchase_order_items\n(purchase_order_id, product_id, quantity_ordered, quantity_received, unit_cost, total_cost)\nVALUES (?, ?, ?, 0, ?, ?)")
        .bind(params.purchase_order_id)
        .bind(params.product_id)
        .bind(&params.quantity_ordered)
        .bind(&params.unit_cost)
        .bind(&params.total_cost)
        .execute(exec)
        .await
}

/// Params for `ListPurchaseOrders` (Go `ListPurchaseOrdersParams`).
pub struct ListPurchaseOrdersParams {
    pub business_id: i64,
    pub limit: i32,
    pub offset: i32,
}

/// `ListPurchaseOrders`: `many`.
pub async fn list_purchase_orders<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListPurchaseOrdersParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, supplier_id, po_number, status, subtotal, tax_amount, total,\n       expected_delivery, received_at, notes, created_by, created_at, updated_at, destination_location_id\nFROM erp_purchase_orders\nWHERE business_id = ?\nORDER BY created_at DESC\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `ListPurchaseOrdersByStatus` (Go `ListPurchaseOrdersByStatusParams`).
pub struct ListPurchaseOrdersByStatusParams {
    pub business_id: i64,
    pub status: String,
    pub limit: i32,
    pub offset: i32,
}

/// `ListPurchaseOrdersByStatus`: `many`.
pub async fn list_purchase_orders_by_status<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListPurchaseOrdersByStatusParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, supplier_id, po_number, status, subtotal, tax_amount, total,\n       expected_delivery, received_at, notes, created_by, created_at, updated_at, destination_location_id\nFROM erp_purchase_orders\nWHERE business_id = ? AND status = ?\nORDER BY created_at DESC\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(&params.status)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `PurchaseOrderExists` (Go `PurchaseOrderExistsParams`).
pub struct PurchaseOrderExistsParams {
    pub business_id: i64,
    pub id: i64,
}

/// `PurchaseOrderExists`: `one`.
pub async fn purchase_order_exists<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &PurchaseOrderExistsParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT EXISTS(SELECT 1 FROM erp_purchase_orders WHERE business_id = ? AND id = ?) AS present")
        .bind(params.business_id)
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `UpdatePurchaseOrderStatus` (Go `UpdatePurchaseOrderStatusParams`).
pub struct UpdatePurchaseOrderStatusParams {
    pub status: String,
    pub business_id: i64,
    pub id: i64,
}

/// `UpdatePurchaseOrderStatus`: `execresult`.
pub async fn update_purchase_order_status<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdatePurchaseOrderStatusParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE erp_purchase_orders\nSET status = ?\nWHERE business_id = ? AND id = ? AND (status IS NULL OR status <> 'received')")
        .bind(&params.status)
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `UpdatePurchaseOrderStatusPartial` (Go `UpdatePurchaseOrderStatusPartialParams`).
pub struct UpdatePurchaseOrderStatusPartialParams {
    pub status: String,
    pub business_id: i64,
    pub id: i64,
}

/// `UpdatePurchaseOrderStatusPartial`: `exec`.
pub async fn update_purchase_order_status_partial<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdatePurchaseOrderStatusPartialParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE erp_purchase_orders SET status = ? WHERE business_id = ? AND id = ?")
        .bind(&params.status)
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `UpdatePurchaseOrderStatusReceived` (Go `UpdatePurchaseOrderStatusReceivedParams`).
pub struct UpdatePurchaseOrderStatusReceivedParams {
    pub status: String,
    pub business_id: i64,
    pub id: i64,
}

/// `UpdatePurchaseOrderStatusReceived`: `execresult`.
pub async fn update_purchase_order_status_received<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdatePurchaseOrderStatusReceivedParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE erp_purchase_orders SET status = ?, received_at = NOW() WHERE business_id = ? AND id = ?")
        .bind(&params.status)
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}
