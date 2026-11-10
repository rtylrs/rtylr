//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `AdvanceOrderSequence` (Go `AdvanceOrderSequenceParams`).
pub struct AdvanceOrderSequenceParams {
    pub last_sequence: i64,
    pub business_id: i64,
    pub prefix: String,
}

/// `AdvanceOrderSequence`: `exec`.
pub async fn advance_order_sequence<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &AdvanceOrderSequenceParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query(
        "UPDATE pos_order_sequences SET last_sequence = ? WHERE business_id = ? AND prefix = ?",
    )
    .bind(params.last_sequence)
    .bind(params.business_id)
    .bind(&params.prefix)
    .execute(exec)
    .await
}

/// Params for `ApplyDiscount` (Go `ApplyDiscountParams`).
pub struct ApplyDiscountParams {
    pub discount_amount: Option<String>,
    pub total: Option<String>,
    pub concat: String,
    pub id: i64,
    pub business_id: i64,
}

/// `ApplyDiscount`: `exec`.
pub async fn apply_discount<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ApplyDiscountParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE pos_orders SET discount_amount = ?, total = ?, notes = CONCAT(COALESCE(notes,''), ?)\nWHERE id = ? AND business_id = ?")
        .bind(&params.discount_amount)
        .bind(&params.total)
        .bind(&params.concat)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `CancelOrder` (Go `CancelOrderParams`).
pub struct CancelOrderParams {
    pub id: i64,
    pub business_id: i64,
}

/// `CancelOrder`: `exec`.
pub async fn cancel_order<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CancelOrderParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE pos_orders SET status = 'cancelled' WHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `CloseSession` (Go `CloseSessionParams`).
pub struct CloseSessionParams {
    pub closing_cash: Option<String>,
    pub expected_cash: Option<String>,
    pub cash_difference: Option<String>,
    pub notes: Option<String>,
    pub business_id: i64,
    pub id: i64,
}

/// `CloseSession`: `exec`.
pub async fn close_session<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CloseSessionParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE pos_sessions\nSET status = 'closed', closing_cash = ?, expected_cash = ?, cash_difference = ?,\n    closed_at = NOW(), notes = ?\nWHERE business_id = ? AND id = ? AND status = 'open'")
        .bind(&params.closing_cash)
        .bind(&params.expected_cash)
        .bind(&params.cash_difference)
        .bind(&params.notes)
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `CountOrdersCreatedBetween` (Go `CountOrdersCreatedBetweenParams`).
pub struct CountOrdersCreatedBetweenParams {
    pub business_id: i64,
    pub date_from: Option<chrono::DateTime<chrono::Utc>>,
    pub date_to: Option<chrono::DateTime<chrono::Utc>>,
}

/// `CountOrdersCreatedBetween`: `one`.
pub async fn count_orders_created_between<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountOrdersCreatedBetweenParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*)\nFROM pos_orders\nWHERE business_id = ?\n  AND created_at >= ?\n  AND created_at <= ?")
        .bind(params.business_id)
        .bind(params.date_from)
        .bind(params.date_to)
        .fetch_optional(exec)
        .await
}

/// Params for `CreateLocation` (Go `CreateLocationParams`).
pub struct CreateLocationParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub name: String,
    pub address: Option<String>,
    pub phone: Option<String>,
}

/// `CreateLocation`: `execresult`.
pub async fn create_location<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateLocationParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO pos_locations (owner_id, business_id, name, address, phone) VALUES (?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(&params.name)
        .bind(&params.address)
        .bind(&params.phone)
        .execute(exec)
        .await
}

/// Params for `CreateOrder` (Go `CreateOrderParams`).
pub struct CreateOrderParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub session_id: Option<i64>,
    pub table_id: Option<i64>,
    pub location_id: Option<i64>,
    pub order_number: String,
    pub r#type: String,
    pub customer_id: Option<i64>,
    pub customer_name: Option<String>,
    pub staff_id: Option<i64>,
    pub subtotal: Option<String>,
    pub tax_amount: Option<String>,
    pub total: Option<String>,
    pub payment_method: String,
    pub notes: Option<String>,
    pub kitchen_notes: Option<String>,
}

/// `CreateOrder`: `execresult`.
pub async fn create_order<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateOrderParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO pos_orders\n(owner_id, business_id, session_id, table_id, location_id, order_number, type, status,\n customer_id, customer_name, staff_id, subtotal, tax_amount, total, payment_method, notes, kitchen_notes)\nVALUES (?, ?, ?, ?, ?, ?, ?, 'pending', ?, ?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.session_id)
        .bind(params.table_id)
        .bind(params.location_id)
        .bind(&params.order_number)
        .bind(&params.r#type)
        .bind(params.customer_id)
        .bind(&params.customer_name)
        .bind(params.staff_id)
        .bind(&params.subtotal)
        .bind(&params.tax_amount)
        .bind(&params.total)
        .bind(&params.payment_method)
        .bind(&params.notes)
        .bind(&params.kitchen_notes)
        .execute(exec)
        .await
}

/// Params for `CreateOrderItem` (Go `CreateOrderItemParams`).
pub struct CreateOrderItemParams {
    pub order_id: i64,
    pub menu_item_id: Option<i64>,
    pub name: String,
    pub sku: Option<String>,
    pub quantity: String,
    pub unit_price: String,
    pub discount_pct: Option<String>,
    pub discount_amount: Option<String>,
    pub tax_rate: Option<String>,
    pub total: String,
    pub notes: Option<String>,
    pub modifiers: String,
    pub combo_id: Option<i64>,
}

/// `CreateOrderItem`: `execresult`.
pub async fn create_order_item<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateOrderItemParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO pos_order_items\n(order_id, menu_item_id, name, sku, quantity, unit_price,\n discount_pct, discount_amount, tax_rate, total, notes, modifiers, combo_id)\nVALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(params.order_id)
        .bind(params.menu_item_id)
        .bind(&params.name)
        .bind(&params.sku)
        .bind(&params.quantity)
        .bind(&params.unit_price)
        .bind(&params.discount_pct)
        .bind(&params.discount_amount)
        .bind(&params.tax_rate)
        .bind(&params.total)
        .bind(&params.notes)
        .bind(&params.modifiers)
        .bind(params.combo_id)
        .execute(exec)
        .await
}

/// Params for `CreatePosTable` (Go `CreatePosTableParams`).
pub struct CreatePosTableParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub location_id: Option<i64>,
    pub name: String,
    pub capacity: Option<i32>,
    pub qr_code_url: Option<String>,
}

/// `CreatePosTable`: `execresult`.
pub async fn create_pos_table<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreatePosTableParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO pos_tables (owner_id, business_id, location_id, name, capacity, qr_code_url)\nVALUES (?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.location_id)
        .bind(&params.name)
        .bind(params.capacity)
        .bind(&params.qr_code_url)
        .execute(exec)
        .await
}

/// Params for `CreatePublicOrder` (Go `CreatePublicOrderParams`).
pub struct CreatePublicOrderParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub order_number: String,
    pub customer_id: Option<i64>,
    pub customer_name: Option<String>,
    pub customer_phone: Option<String>,
    pub subtotal: Option<String>,
    pub tax_amount: Option<String>,
    pub total: Option<String>,
    pub notes: Option<String>,
}

/// `CreatePublicOrder`: `execresult`.
pub async fn create_public_order<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreatePublicOrderParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO pos_orders\n(owner_id, business_id, order_number, type, status, customer_id, customer_name, customer_phone,\n subtotal, tax_amount, total, payment_status, notes, created_at, updated_at)\nVALUES (?, ?, ?, 'online', 'pending', ?, ?, ?, ?, ?, ?, 'unpaid', ?, NOW(), NOW())")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(&params.order_number)
        .bind(params.customer_id)
        .bind(&params.customer_name)
        .bind(&params.customer_phone)
        .bind(&params.subtotal)
        .bind(&params.tax_amount)
        .bind(&params.total)
        .bind(&params.notes)
        .execute(exec)
        .await
}

/// Params for `CreatePublicOrderItem` (Go `CreatePublicOrderItemParams`).
pub struct CreatePublicOrderItemParams {
    pub order_id: i64,
    pub menu_item_id: Option<i64>,
    pub name: String,
    pub quantity: String,
    pub unit_price: String,
    pub total: String,
    pub notes: Option<String>,
}

/// `CreatePublicOrderItem`: `exec`.
pub async fn create_public_order_item<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreatePublicOrderItemParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO pos_order_items\n(order_id, menu_item_id, name, quantity, unit_price, total, notes, status, created_at)\nVALUES (?, ?, ?, ?, ?, ?, ?, 'pending', NOW())")
        .bind(params.order_id)
        .bind(params.menu_item_id)
        .bind(&params.name)
        .bind(&params.quantity)
        .bind(&params.unit_price)
        .bind(&params.total)
        .bind(&params.notes)
        .execute(exec)
        .await
}

/// Params for `CreateRefund` (Go `CreateRefundParams`).
pub struct CreateRefundParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub order_id: i64,
    pub amount: String,
    pub reason: Option<String>,
    pub refunded_by: Option<i64>,
}

/// `CreateRefund`: `execresult`.
pub async fn create_refund<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateRefundParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO pos_refunds (owner_id, business_id, order_id, amount, reason, refunded_by)\nVALUES (?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.order_id)
        .bind(&params.amount)
        .bind(&params.reason)
        .bind(params.refunded_by)
        .execute(exec)
        .await
}

/// Params for `DeletePosTable` (Go `DeletePosTableParams`).
pub struct DeletePosTableParams {
    pub business_id: i64,
    pub id: i64,
}

/// `DeletePosTable`: `exec`.
pub async fn delete_pos_table<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeletePosTableParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM pos_tables WHERE business_id = ? AND id = ?")
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `EnsureOrderSequence` (Go `EnsureOrderSequenceParams`).
pub struct EnsureOrderSequenceParams {
    pub business_id: i64,
    pub prefix: String,
}

/// `EnsureOrderSequence`: `exec`.
pub async fn ensure_order_sequence<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &EnsureOrderSequenceParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO pos_order_sequences (business_id, prefix, last_sequence) VALUES (?, ?, 0)\nON DUPLICATE KEY UPDATE last_sequence = last_sequence")
        .bind(params.business_id)
        .bind(&params.prefix)
        .execute(exec)
        .await
}

/// Params for `FreeTable` (Go `FreeTableParams`).
pub struct FreeTableParams {
    pub id: i64,
    pub business_id: i64,
}

/// `FreeTable`: `exec`.
pub async fn free_table<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &FreeTableParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE pos_tables SET status = 'available', current_session_order_id = NULL WHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `FreeTableByID` (Go `FreeTableByIDParams`).
pub struct FreeTableByIDParams {
    pub id: i64,
    pub business_id: i64,
}

/// `FreeTableByID`: `exec`.
pub async fn free_table_by_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &FreeTableByIDParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE pos_tables SET status = 'available' WHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `GetActiveOrderItemsTotal` (Go `GetActiveOrderItemsTotalParams`).
pub struct GetActiveOrderItemsTotalParams {
    pub order_id: i64,
}

/// `GetActiveOrderItemsTotal`: `one`.
pub async fn get_active_order_items_total<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetActiveOrderItemsTotalParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COALESCE(SUM((unit_price * quantity) - COALESCE(discount_amount, 0)), 0) AS subtotal,\n       COALESCE(SUM(((unit_price * quantity) - COALESCE(discount_amount, 0)) * COALESCE(tax_rate, 0) / 100), 0) AS tax_amount\nFROM pos_order_items\nWHERE order_id = ?\n  AND status != 'cancelled'")
        .bind(params.order_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetCurrentOpenSession` (Go `GetCurrentOpenSessionParams`).
pub struct GetCurrentOpenSessionParams {
    pub business_id: i64,
    pub staff_id: i64,
}

/// `GetCurrentOpenSession`: `one`.
pub async fn get_current_open_session<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetCurrentOpenSessionParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, location_id, staff_id, status,\n       opening_cash, closing_cash, expected_cash, cash_difference,\n       total_sales, total_orders, total_cash_sales, total_card_sales,\n       total_online_sales, total_refunds, opened_at, closed_at, notes,\n       terminal_id\nFROM pos_sessions\nWHERE business_id = ? AND staff_id = ? AND status = 'open'\nLIMIT 1")
        .bind(params.business_id)
        .bind(params.staff_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetDashboardSalesToday` (Go `GetDashboardSalesTodayParams`).
pub struct GetDashboardSalesTodayParams {
    pub business_id: i64,
}

/// `GetDashboardSalesToday`: `one`.
pub async fn get_dashboard_sales_today<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetDashboardSalesTodayParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT\n    COALESCE(SUM(total), 0) AS total_sales,\n    COUNT(*) AS order_count,\n    COALESCE(AVG(total), 0) AS avg_order\nFROM pos_orders\nWHERE business_id = ?\n  AND payment_status = 'paid'\n  AND paid_at >= CURDATE()\n  AND paid_at < DATE_ADD(CURDATE(), INTERVAL 1 DAY)")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetOrder` (Go `GetOrderParams`).
pub struct GetOrderParams {
    pub business_id: i64,
    pub id: i64,
}

/// `GetOrder`: `one`.
pub async fn get_order<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetOrderParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, session_id, table_id, location_id,\n       order_number, type, status, customer_id, customer_name,\n       staff_id, subtotal, discount_amount, tax_amount, total,\n       payment_method, payment_status, amount_paid, change_due,\n       notes, kitchen_notes, created_at, updated_at, paid_at,\n       terminal_id, branch_id, tip_amount, void_reason, voided_by, voided_at, exchange_order_id\nFROM pos_orders\nWHERE business_id = ? AND id = ?")
        .bind(params.business_id)
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetOrderForDiscount` (Go `GetOrderForDiscountParams`).
pub struct GetOrderForDiscountParams {
    pub business_id: i64,
    pub id: i64,
}

/// `GetOrderForDiscount`: `one`.
pub async fn get_order_for_discount<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetOrderForDiscountParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query(
        "SELECT payment_status\nFROM pos_orders\nWHERE business_id = ? AND id = ?\nFOR UPDATE",
    )
    .bind(params.business_id)
    .bind(params.id)
    .fetch_optional(exec)
    .await
}

/// Params for `GetOrderForPayment` (Go `GetOrderForPaymentParams`).
pub struct GetOrderForPaymentParams {
    pub business_id: i64,
    pub id: i64,
}

/// `GetOrderForPayment`: `one`.
pub async fn get_order_for_payment<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetOrderForPaymentParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT total, table_id, type, payment_status FROM pos_orders WHERE business_id = ? AND id = ? FOR UPDATE")
        .bind(params.business_id)
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetOrderForRefund` (Go `GetOrderForRefundParams`).
pub struct GetOrderForRefundParams {
    pub business_id: i64,
    pub id: i64,
}

/// `GetOrderForRefund`: `one`.
pub async fn get_order_for_refund<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetOrderForRefundParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT total, amount_paid, change_due, payment_status FROM pos_orders WHERE business_id = ? AND id = ? FOR UPDATE")
        .bind(params.business_id)
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetOrderIDByPaddleTransaction` (Go `GetOrderIDByPaddleTransactionParams`).
pub struct GetOrderIDByPaddleTransactionParams {
    pub paddle_transaction_id: Option<String>,
}

/// `GetOrderIDByPaddleTransaction`: `one`.
pub async fn get_order_id_by_paddle_transaction<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetOrderIDByPaddleTransactionParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id FROM pos_orders WHERE paddle_transaction_id = ? LIMIT 1")
        .bind(&params.paddle_transaction_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetOrderItemStatus` (Go `GetOrderItemStatusParams`).
pub struct GetOrderItemStatusParams {
    pub id: i64,
    pub order_id: i64,
}

/// `GetOrderItemStatus`: `one`.
pub async fn get_order_item_status<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetOrderItemStatusParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT status FROM pos_order_items WHERE id = ? AND order_id = ?")
        .bind(params.id)
        .bind(params.order_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetOrderItems` (Go `GetOrderItemsParams`).
pub struct GetOrderItemsParams {
    pub order_id: i64,
}

/// `GetOrderItems`: `many`.
pub async fn get_order_items<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetOrderItemsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, order_id, menu_item_id, name, sku,\n       quantity, unit_price, discount_pct, discount_amount, tax_rate, total,\n       notes, status, created_at, COALESCE(modifiers, JSON_ARRAY()) AS modifiers, combo_id, variant_name\nFROM pos_order_items\nWHERE order_id = ?")
        .bind(params.order_id)
        .execute(exec)
        .await
}

/// Params for `GetOrderItemsForOrders` (Go `GetOrderItemsForOrdersParams`).
pub struct GetOrderItemsForOrdersParams {
    pub order_ids: Vec<i64>,
}

/// `GetOrderItemsForOrders`: `many`.
pub async fn get_order_items_for_orders<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetOrderItemsForOrdersParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    let mut sql_template: String = "SELECT id, order_id, COALESCE(menu_item_id,0) AS menu_item_id, name, COALESCE(sku,'') AS sku,\n       quantity, unit_price, COALESCE(discount_pct,0) AS discount_pct, COALESCE(discount_amount,0) AS discount_amount,\n       COALESCE(tax_rate,0) AS tax_rate, total, COALESCE(notes,'') AS notes, COALESCE(status,'pending') AS status,\n       created_at\nFROM pos_order_items\nWHERE order_id IN ({SLICE:order_ids})\nORDER BY created_at ASC, id ASC".to_string();

    if params.order_ids.is_empty() {
        sql_template = sql_template.replace("{SLICE:order_ids}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:order_ids}",
            &std::iter::repeat_n("?", params.order_ids.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    let query = sqlx::query(&sql_template);
    let mut query = query;
    for value in &params.order_ids {
        query = query.bind(value);
    }
    query.execute(exec).await
}

/// Params for `GetOrderLocationID` (Go `GetOrderLocationIDParams`).
pub struct GetOrderLocationIDParams {
    pub business_id: i64,
    pub id: i64,
}

/// `GetOrderLocationID`: `one`.
pub async fn get_order_location_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetOrderLocationIDParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COALESCE(location_id, 0) FROM pos_orders WHERE business_id = ? AND id = ?")
        .bind(params.business_id)
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetOrderSessionID` (Go `GetOrderSessionIDParams`).
pub struct GetOrderSessionIDParams {
    pub id: i64,
}

/// `GetOrderSessionID`: `one`.
pub async fn get_order_session_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetOrderSessionIDParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT session_id FROM pos_orders WHERE id = ?")
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetOrderSubtotal` (Go `GetOrderSubtotalParams`).
pub struct GetOrderSubtotalParams {
    pub business_id: i64,
    pub id: i64,
}

/// `GetOrderSubtotal`: `one`.
pub async fn get_order_subtotal<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetOrderSubtotalParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT o.id,\n       COALESCE(o.discount_amount, 0) AS existing_discount,\n       COALESCE(SUM((i.unit_price * i.quantity) - COALESCE(i.discount_amount, 0)), 0) AS subtotal,\n       COALESCE(SUM(((i.unit_price * i.quantity) - COALESCE(i.discount_amount, 0)) * COALESCE(i.tax_rate, 0) / 100), 0) AS tax_amount\nFROM pos_orders o\nLEFT JOIN pos_order_items i ON i.order_id = o.id AND i.status != 'cancelled'\nWHERE o.business_id = ? AND o.id = ?\nGROUP BY o.id")
        .bind(params.business_id)
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetPOSComboForOrder` (Go `GetPOSComboForOrderParams`).
pub struct GetPOSComboForOrderParams {
    pub business_id: i64,
    pub id: i64,
}

/// `GetPOSComboForOrder`: `one`.
pub async fn get_pos_combo_for_order<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetPOSComboForOrderParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT name, price\nFROM pos_combos\nWHERE business_id = ? AND id = ? AND is_active = TRUE")
        .bind(params.business_id)
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetPOSModifierTotal` (Go `GetPOSModifierTotalParams`).
pub struct GetPOSModifierTotalParams {
    pub business_id: i64,
    pub modifier_ids: Vec<i64>,
}

/// `GetPOSModifierTotal`: `one`.
pub async fn get_pos_modifier_total<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetPOSModifierTotalParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    let mut sql_template: String = "SELECT COALESCE(SUM(price_delta), 0) AS price_delta, COUNT(*) AS modifier_count\nFROM pos_modifiers\nWHERE business_id = ? AND id IN ({SLICE:modifier_ids})".to_string();

    if params.modifier_ids.is_empty() {
        sql_template = sql_template.replace("{SLICE:modifier_ids}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:modifier_ids}",
            &std::iter::repeat_n("?", params.modifier_ids.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    let query = sqlx::query(&sql_template);
    let mut query = query.bind(params.business_id);
    for value in &params.modifier_ids {
        query = query.bind(value);
    }
    query.fetch_optional(exec).await
}

/// Params for `GetPOSUserCapabilities` (Go `GetPOSUserCapabilitiesParams`).
pub struct GetPOSUserCapabilitiesParams {
    pub business_id: i64,
    pub user_id: i64,
    pub location_id: i64,
    pub column4: String,
    pub location_id_2: i64,
}

/// `GetPOSUserCapabilities`: `one`.
pub async fn get_pos_user_capabilities<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetPOSUserCapabilitiesParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT can_apply_discount, max_discount_pct\nFROM pos_users\nWHERE business_id = ?\n  AND user_id = ?\n  AND is_active = TRUE\n  AND (location_id = ? OR ? = 0)\nORDER BY CASE WHEN location_id = ? THEN 0 ELSE 1 END, id\nLIMIT 1")
        .bind(params.business_id)
        .bind(params.user_id)
        .bind(params.location_id)
        .bind(&params.column4)
        .bind(params.location_id_2)
        .fetch_optional(exec)
        .await
}

/// Params for `GetPublicOrderNumberStatus` (Go `GetPublicOrderNumberStatusParams`).
pub struct GetPublicOrderNumberStatusParams {
    pub business_id: i64,
    pub id: i64,
}

/// `GetPublicOrderNumberStatus`: `one`.
pub async fn get_public_order_number_status<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetPublicOrderNumberStatusParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query(
        "SELECT order_number, status FROM pos_orders WHERE business_id = ? AND id = ? LIMIT 1",
    )
    .bind(params.business_id)
    .bind(params.id)
    .fetch_optional(exec)
    .await
}

/// Params for `GetSessionByID` (Go `GetSessionByIDParams`).
pub struct GetSessionByIDParams {
    pub business_id: i64,
    pub id: i64,
}

/// `GetSessionByID`: `one`.
pub async fn get_session_by_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetSessionByIDParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, location_id, staff_id, status,\n       opening_cash, closing_cash, expected_cash, cash_difference,\n       total_sales, total_orders, total_cash_sales, total_card_sales,\n       total_online_sales, total_refunds, opened_at, closed_at, notes,\n       terminal_id\nFROM pos_sessions\nWHERE business_id = ? AND id = ?")
        .bind(params.business_id)
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetSessionTotals` (Go `GetSessionTotalsParams`).
pub struct GetSessionTotalsParams {
    pub session_id: Option<i64>,
}

/// `GetSessionTotals`: `one`.
pub async fn get_session_totals<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetSessionTotalsParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT\n    COALESCE(SUM(total), 0) AS total_sales,\n    COALESCE(SUM(CASE WHEN payment_method = 'cash' THEN total ELSE 0 END), 0) AS cash_sales,\n    COALESCE(SUM(CASE WHEN payment_method = 'card' THEN total ELSE 0 END), 0) AS card_sales,\n    COALESCE(SUM(CASE WHEN payment_method = 'online' THEN total ELSE 0 END), 0) AS online_sales,\n    COUNT(*) AS total_orders\nFROM pos_orders\nWHERE session_id = ? AND payment_status = 'paid'")
        .bind(params.session_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetTableTurnover` (Go `GetTableTurnoverParams`).
pub struct GetTableTurnoverParams {
    pub business_id: i64,
}

/// `GetTableTurnover`: `one`.
pub async fn get_table_turnover<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetTableTurnoverParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) AS table_count,\n       COALESCE(SUM(CASE WHEN status != 'available' THEN 1 ELSE 0 END), 0) AS turned_tables\nFROM pos_tables WHERE business_id = ?")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `HoldOrder` (Go `HoldOrderParams`).
pub struct HoldOrderParams {
    pub concat: String,
    pub id: i64,
    pub business_id: i64,
}

/// `HoldOrder`: `execresult`.
pub async fn hold_order<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &HoldOrderParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE pos_orders SET status = 'held', notes = CONCAT(COALESCE(notes,''), ?)\nWHERE id = ? AND business_id = ?")
        .bind(&params.concat)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `InsightsBuyerSpendBuckets` (Go `InsightsBuyerSpendBucketsParams`).
pub struct InsightsBuyerSpendBucketsParams {
    pub business_id: i64,
}

/// `InsightsBuyerSpendBuckets`: `many`.
pub async fn insights_buyer_spend_buckets<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsightsBuyerSpendBucketsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("WITH buckets AS (\n  SELECT 0 AS bucket_order, '0-100' AS label, 0 AS lower_minor_units, 10000 AS upper_minor_units\n  UNION ALL SELECT 1, '100-500', 10000, 50000\n  UNION ALL SELECT 2, '500-1k', 50000, 100000\n  UNION ALL SELECT 3, '1k-5k', 100000, 500000\n  UNION ALL SELECT 4, '5k-10k', 500000, 1000000\n  UNION ALL SELECT 5, '10k+', 1000000, 0\n),\ncustomer_spend AS (\n  SELECT customer_id, COALESCE(SUM(total), 0) * 100 AS spend_minor_units\n  FROM pos_orders\n  WHERE business_id = ?\n    AND customer_id IS NOT NULL\n    AND payment_status = 'paid'\n  GROUP BY customer_id\n)\nSELECT b.label, b.lower_minor_units, b.upper_minor_units, COUNT(cs.customer_id) AS customer_count\nFROM buckets b\nLEFT JOIN customer_spend cs\n  ON cs.spend_minor_units >= b.lower_minor_units\n AND (b.upper_minor_units = 0 OR cs.spend_minor_units < b.upper_minor_units)\nGROUP BY b.bucket_order, b.label, b.lower_minor_units, b.upper_minor_units\nORDER BY b.bucket_order")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `InsightsCountBuyers` (Go `InsightsCountBuyersParams`).
pub struct InsightsCountBuyersParams {
    pub business_id: i64,
    pub date_from: Option<chrono::DateTime<chrono::Utc>>,
    pub date_to: chrono::NaiveDate,
}

/// `InsightsCountBuyers`: `one`.
pub async fn insights_count_buyers<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsightsCountBuyersParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(DISTINCT customer_id) AS count\nFROM pos_orders\nWHERE business_id = ?\n  AND customer_id IS NOT NULL\n  AND payment_status = 'paid'\n  AND created_at >= ?\n  AND created_at < DATE_ADD(?, INTERVAL 1 DAY)")
        .bind(params.business_id)
        .bind(params.date_from)
        .bind(params.date_to)
        .fetch_optional(exec)
        .await
}

/// Params for `InsightsCountLapsedBuyers` (Go `InsightsCountLapsedBuyersParams`).
pub struct InsightsCountLapsedBuyersParams {
    pub business_id: i64,
    pub date_to: chrono::NaiveDate,
    pub date_from: Option<chrono::DateTime<chrono::Utc>>,
}

/// `InsightsCountLapsedBuyers`: `one`.
pub async fn insights_count_lapsed_buyers<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsightsCountLapsedBuyersParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) AS count\nFROM (\n  SELECT customer_id\n  FROM pos_orders\n  WHERE business_id = ?\n    AND customer_id IS NOT NULL\n    AND payment_status = 'paid'\n    AND created_at < DATE_ADD(?, INTERVAL 1 DAY)\n  GROUP BY customer_id\n  HAVING MAX(created_at) < ?\n) AS lapsed")
        .bind(params.business_id)
        .bind(params.date_to)
        .bind(params.date_from)
        .fetch_optional(exec)
        .await
}

/// Params for `InsightsCountOrdersToday` (Go `InsightsCountOrdersTodayParams`).
pub struct InsightsCountOrdersTodayParams {
    pub business_id: i64,
}

/// `InsightsCountOrdersToday`: `one`.
pub async fn insights_count_orders_today<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsightsCountOrdersTodayParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM pos_orders\nWHERE business_id = ?\n  AND created_at >= CURDATE()\n  AND created_at < CURDATE() + INTERVAL 1 DAY")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `InsightsLifetimeProductSales` (Go `InsightsLifetimeProductSalesParams`).
pub struct InsightsLifetimeProductSalesParams {
    pub business_id: i64,
}

/// `InsightsLifetimeProductSales`: `many`.
pub async fn insights_lifetime_product_sales<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsightsLifetimeProductSalesParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT oi.menu_item_id AS product_id,\n       COALESCE(SUM(oi.quantity), 0) AS units,\n       CAST(COALESCE(SUM(oi.total), 0) * 100 AS SIGNED) AS revenue_minor_units\nFROM pos_order_items oi\nJOIN pos_orders o ON o.id = oi.order_id\nWHERE o.business_id = ?\n  AND oi.menu_item_id IS NOT NULL\nGROUP BY oi.menu_item_id\nHAVING SUM(oi.quantity) > 0")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `InsightsNetRevenue` (Go `InsightsNetRevenueParams`).
pub struct InsightsNetRevenueParams {
    pub business_id: i64,
    pub date_from: Option<chrono::DateTime<chrono::Utc>>,
    pub date_to: Option<chrono::DateTime<chrono::Utc>>,
}

/// `InsightsNetRevenue`: `one`.
pub async fn insights_net_revenue<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsightsNetRevenueParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COALESCE(SUM(total), 0) AS total\nFROM (\n    SELECT GREATEST(o.total - COALESCE(SUM(r.amount), 0), 0) AS total\n    FROM pos_orders o\n    LEFT JOIN pos_refunds r ON r.order_id = o.id AND r.business_id = o.business_id\n    WHERE o.business_id = ?\n      AND o.payment_status IN ('paid', 'refunded')\n      AND o.paid_at >= ?\n      AND o.paid_at <= ?\n    GROUP BY o.id, o.total\n) AS net_pos_orders")
        .bind(params.business_id)
        .bind(params.date_from)
        .bind(params.date_to)
        .fetch_optional(exec)
        .await
}

/// Params for `InsightsPOSLastChange` (Go `InsightsPOSLastChangeParams`).
pub struct InsightsPOSLastChangeParams {
    pub business_id: i64,
}

/// `InsightsPOSLastChange`: `one`.
pub async fn insights_pos_last_change<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsightsPOSLastChangeParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT CAST(GREATEST(\n  COALESCE((SELECT MAX(o.updated_at) FROM pos_orders o WHERE o.business_id = ?), '1000-01-01'),\n  COALESCE((SELECT MAX(r.refunded_at) FROM pos_refunds r WHERE r.business_id = ?), '1000-01-01')\n) AS DATETIME) AS last_change")
        .bind(params.business_id)
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `InsightsPaidProductSales` (Go `InsightsPaidProductSalesParams`).
pub struct InsightsPaidProductSalesParams {
    pub business_id: i64,
    pub date_from: Option<chrono::DateTime<chrono::Utc>>,
    pub date_to: Option<chrono::DateTime<chrono::Utc>>,
}

/// `InsightsPaidProductSales`: `many`.
pub async fn insights_paid_product_sales<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsightsPaidProductSalesParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT oi.menu_item_id AS product_id,\n       COALESCE(SUM(oi.quantity), 0) AS units,\n       COALESCE(SUM(oi.total), 0)    AS revenue\nFROM pos_order_items oi\nJOIN pos_orders o ON o.id = oi.order_id\nWHERE o.business_id = ?\n  AND o.payment_status = 'paid'\n  AND oi.status <> 'cancelled'\n  AND oi.menu_item_id IS NOT NULL\n  AND o.created_at >= ?\n  AND (? IS NULL OR o.created_at <= ?)\nGROUP BY oi.menu_item_id")
        .bind(params.business_id)
        .bind(params.date_from)
        .bind(params.date_to)
        .bind(params.date_to)
        .execute(exec)
        .await
}

/// Params for `InsightsProductsSoldSince` (Go `InsightsProductsSoldSinceParams`).
pub struct InsightsProductsSoldSinceParams {
    pub business_id: i64,
    pub created_at: Option<chrono::DateTime<chrono::Utc>>,
}

/// `InsightsProductsSoldSince`: `many`.
pub async fn insights_products_sold_since<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsightsProductsSoldSinceParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT DISTINCT oi.menu_item_id\nFROM pos_order_items oi\nJOIN pos_orders o ON o.id = oi.order_id\nWHERE o.business_id = ?\n  AND o.created_at >= ?\n  AND oi.menu_item_id IS NOT NULL")
        .bind(params.business_id)
        .bind(params.created_at)
        .execute(exec)
        .await
}

/// Params for `InsightsSumRevenueToday` (Go `InsightsSumRevenueTodayParams`).
pub struct InsightsSumRevenueTodayParams {
    pub business_id: i64,
}

/// `InsightsSumRevenueToday`: `one`.
pub async fn insights_sum_revenue_today<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsightsSumRevenueTodayParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COALESCE(SUM(total), 0) FROM pos_orders\nWHERE business_id = ?\n  AND created_at >= CURDATE()\n  AND created_at < CURDATE() + INTERVAL 1 DAY")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `InsightsTopBuyers` (Go `InsightsTopBuyersParams`).
pub struct InsightsTopBuyersParams {
    pub business_id: i64,
    pub date_from: Option<chrono::DateTime<chrono::Utc>>,
    pub date_to: chrono::NaiveDate,
    pub by_order_count: i64,
    pub direction: i64,
    pub limit: i32,
    pub offset: i32,
}

/// `InsightsTopBuyers`: `many`.
pub async fn insights_top_buyers<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsightsTopBuyersParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT\n  customer_id,\n  COALESCE(MAX(customer_name), '') AS customer_name,\n  SUM(total) AS spend,\n  COUNT(id) AS orders,\n  MAX(created_at) AS last_order_at\nFROM pos_orders\nWHERE business_id = ?\n  AND customer_id IS NOT NULL\n  AND payment_status = 'paid'\n  AND created_at >= ?\n  AND created_at < DATE_ADD(?, INTERVAL 1 DAY)\nGROUP BY customer_id\nORDER BY\n  (CASE WHEN ? THEN COUNT(id) ELSE SUM(total) END) * ? DESC,\n  customer_id ASC\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(params.date_from)
        .bind(params.date_to)
        .bind(params.by_order_count)
        .bind(params.direction)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `IsPaddleEventProcessed` (Go `IsPaddleEventProcessedParams`).
pub struct IsPaddleEventProcessedParams {
    pub transaction_id: String,
}

/// `IsPaddleEventProcessed`: `one`.
pub async fn is_paddle_event_processed<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &IsPaddleEventProcessedParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM pos_paddle_events_processed WHERE transaction_id = ?")
        .bind(&params.transaction_id)
        .fetch_optional(exec)
        .await
}

/// Params for `ListLocationsByOwner` (Go `ListLocationsByOwnerParams`).
pub struct ListLocationsByOwnerParams {
    pub business_id: i64,
}

/// `ListLocationsByOwner`: `many`.
pub async fn list_locations_by_owner<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListLocationsByOwnerParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, name, address, phone, is_active, created_at\nFROM pos_locations WHERE business_id = ? ORDER BY name ASC LIMIT 500")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `ListPaidOrderProductLines` (Go `ListPaidOrderProductLinesParams`).
pub struct ListPaidOrderProductLinesParams {
    pub business_id: i64,
    pub id: i64,
}

/// `ListPaidOrderProductLines`: `many`.
pub async fn list_paid_order_product_lines<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListPaidOrderProductLinesParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT oi.menu_item_id, oi.quantity\nFROM pos_order_items oi\nJOIN pos_orders o ON o.id = oi.order_id\nWHERE o.business_id = ? AND o.id = ? AND oi.menu_item_id IS NOT NULL\nORDER BY oi.id")
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `ListTablesByOwner` (Go `ListTablesByOwnerParams`).
pub struct ListTablesByOwnerParams {
    pub business_id: i64,
    pub location_id: Option<i64>,
    pub status: String,
}

/// `ListTablesByOwner`: `many`.
pub async fn list_tables_by_owner<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListTablesByOwnerParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, location_id, name, capacity, status,\n       current_session_order_id, qr_code_url, created_at, updated_at\nFROM pos_tables\nWHERE business_id = ?\n  AND (? = 0 OR location_id = ?)\n  AND (? = '' OR status = ?)\nORDER BY name ASC, id ASC")
        .bind(params.business_id)
        .bind(params.location_id)
        .bind(params.location_id)
        .bind(&params.status)
        .bind(&params.status)
        .execute(exec)
        .await
}

/// Params for `LockOrderSequence` (Go `LockOrderSequenceParams`).
pub struct LockOrderSequenceParams {
    pub business_id: i64,
    pub prefix: String,
}

/// `LockOrderSequence`: `one`.
pub async fn lock_order_sequence<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &LockOrderSequenceParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT last_sequence FROM pos_order_sequences WHERE business_id = ? AND prefix = ? FOR UPDATE")
        .bind(params.business_id)
        .bind(&params.prefix)
        .fetch_optional(exec)
        .await
}

/// Params for `MarkOrderPaid` (Go `MarkOrderPaidParams`).
pub struct MarkOrderPaidParams {
    pub payment_method: String,
    pub amount_paid: Option<String>,
    pub change_due: Option<String>,
    pub paid_at: Option<chrono::DateTime<chrono::Utc>>,
    pub business_id: i64,
    pub id: i64,
}

/// `MarkOrderPaid`: `execresult`.
pub async fn mark_order_paid<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &MarkOrderPaidParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE pos_orders\nSET status = 'paid', payment_status = 'paid', payment_method = ?,\n    amount_paid = ?, change_due = ?, paid_at = ?\nWHERE business_id = ? AND id = ? AND payment_status <> 'paid'")
        .bind(&params.payment_method)
        .bind(&params.amount_paid)
        .bind(&params.change_due)
        .bind(params.paid_at)
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `MarkOrderRefunded` (Go `MarkOrderRefundedParams`).
pub struct MarkOrderRefundedParams {
    pub business_id: i64,
    pub id: i64,
}

/// `MarkOrderRefunded`: `exec`.
pub async fn mark_order_refunded<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &MarkOrderRefundedParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE pos_orders SET status = 'refunded', payment_status = 'refunded'\nWHERE business_id = ? AND id = ?")
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `MaxOrderNumberSequence` (Go `MaxOrderNumberSequenceParams`).
pub struct MaxOrderNumberSequenceParams {
    pub business_id: i64,
    pub prefix: String,
}

/// `MaxOrderNumberSequence`: `one`.
pub async fn max_order_number_sequence<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &MaxOrderNumberSequenceParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT CAST(COALESCE(MAX(CAST(SUBSTRING_INDEX(order_number, '-', -1) AS UNSIGNED)), 0) AS UNSIGNED) AS sequence\nFROM pos_orders\nWHERE business_id = ?\n  AND order_number LIKE CONCAT(?, '%')\n  AND order_number REGEXP '^ORD-[0-9]{8}-[0-9]+$'")
        .bind(params.business_id)
        .bind(&params.prefix)
        .fetch_optional(exec)
        .await
}

/// Params for `MoveOrderItems` (Go `MoveOrderItemsParams`).
pub struct MoveOrderItemsParams {
    pub order_id: i64,
    pub id: i64,
    pub order_id_2: i64,
}

/// `MoveOrderItems`: `exec`.
pub async fn move_order_items<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &MoveOrderItemsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE pos_order_items SET order_id = ? WHERE id = ? AND order_id = ?")
        .bind(params.order_id)
        .bind(params.id)
        .bind(params.order_id_2)
        .execute(exec)
        .await
}

/// Params for `MoveOrderToTable` (Go `MoveOrderToTableParams`).
pub struct MoveOrderToTableParams {
    pub table_id: Option<i64>,
    pub id: i64,
    pub business_id: i64,
}

/// `MoveOrderToTable`: `exec`.
pub async fn move_order_to_table<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &MoveOrderToTableParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE pos_orders SET table_id = ? WHERE id = ? AND business_id = ?")
        .bind(params.table_id)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `OccupyTableByID` (Go `OccupyTableByIDParams`).
pub struct OccupyTableByIDParams {
    pub id: i64,
    pub business_id: i64,
}

/// `OccupyTableByID`: `execresult`.
pub async fn occupy_table_by_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &OccupyTableByIDParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE pos_tables SET status = 'occupied' WHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `OpenSession` (Go `OpenSessionParams`).
pub struct OpenSessionParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub location_id: Option<i64>,
    pub staff_id: i64,
    pub opening_cash: Option<String>,
}

/// `OpenSession`: `execresult`.
pub async fn open_session<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &OpenSessionParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO pos_sessions (owner_id, business_id, location_id, staff_id, status, opening_cash)\nVALUES (?, ?, ?, ?, 'open', ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.location_id)
        .bind(params.staff_id)
        .bind(&params.opening_cash)
        .execute(exec)
        .await
}

/// Params for `OrderExists` (Go `OrderExistsParams`).
pub struct OrderExistsParams {
    pub business_id: i64,
    pub id: i64,
}

/// `OrderExists`: `one`.
pub async fn order_exists<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &OrderExistsParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM pos_orders WHERE business_id = ? AND id = ?")
        .bind(params.business_id)
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `OrderItemExists` (Go `OrderItemExistsParams`).
pub struct OrderItemExistsParams {
    pub order_id: i64,
    pub id: i64,
}

/// `OrderItemExists`: `one`.
pub async fn order_item_exists<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &OrderItemExistsParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM pos_order_items WHERE order_id = ? AND id = ?")
        .bind(params.order_id)
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `RecordPaddleEventProcessed` (Go `RecordPaddleEventProcessedParams`).
pub struct RecordPaddleEventProcessedParams {
    pub transaction_id: String,
    pub event_type: String,
    pub owner_id: i64,
}

/// `RecordPaddleEventProcessed`: `execresult`.
pub async fn record_paddle_event_processed<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &RecordPaddleEventProcessedParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO pos_paddle_events_processed (transaction_id, event_type, owner_id)\nVALUES (?, ?, ?)")
        .bind(&params.transaction_id)
        .bind(&params.event_type)
        .bind(params.owner_id)
        .execute(exec)
        .await
}

/// Params for `SetOrderPaddleTransactionID` (Go `SetOrderPaddleTransactionIDParams`).
pub struct SetOrderPaddleTransactionIDParams {
    pub paddle_transaction_id: Option<String>,
    pub id: i64,
    pub business_id: i64,
}

/// `SetOrderPaddleTransactionID`: `execresult`.
pub async fn set_order_paddle_transaction_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &SetOrderPaddleTransactionIDParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE pos_orders SET paddle_transaction_id = ?\nWHERE id = ? AND business_id = ?")
        .bind(&params.paddle_transaction_id)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `SetTableOccupied` (Go `SetTableOccupiedParams`).
pub struct SetTableOccupiedParams {
    pub current_session_order_id: Option<i64>,
    pub business_id: i64,
    pub id: i64,
}

/// `SetTableOccupied`: `execresult`.
pub async fn set_table_occupied<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &SetTableOccupiedParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE pos_tables SET status = 'occupied', current_session_order_id = ? WHERE business_id = ? AND id = ?")
        .bind(params.current_session_order_id)
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `SumRefundsForOrder` (Go `SumRefundsForOrderParams`).
pub struct SumRefundsForOrderParams {
    pub business_id: i64,
    pub order_id: i64,
}

/// `SumRefundsForOrder`: `one`.
pub async fn sum_refunds_for_order<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &SumRefundsForOrderParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COALESCE(SUM(amount), 0) AS refunded FROM pos_refunds WHERE business_id = ? AND order_id = ?")
        .bind(params.business_id)
        .bind(params.order_id)
        .fetch_optional(exec)
        .await
}

/// Params for `UpdateOrderItemStatus` (Go `UpdateOrderItemStatusParams`).
pub struct UpdateOrderItemStatusParams {
    pub status: String,
    pub order_id: i64,
    pub id: i64,
}

/// `UpdateOrderItemStatus`: `execresult`.
pub async fn update_order_item_status<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateOrderItemStatusParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE pos_order_items SET status = ? WHERE order_id = ? AND id = ?")
        .bind(&params.status)
        .bind(params.order_id)
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `UpdateOrderStatus` (Go `UpdateOrderStatusParams`).
pub struct UpdateOrderStatusParams {
    pub status: String,
    pub business_id: i64,
    pub id: i64,
}

/// `UpdateOrderStatus`: `execresult`.
pub async fn update_order_status<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateOrderStatusParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE pos_orders SET status = ? WHERE business_id = ? AND id = ?")
        .bind(&params.status)
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `UpdateOrderTotal` (Go `UpdateOrderTotalParams`).
pub struct UpdateOrderTotalParams {
    pub subtotal: Option<String>,
    pub tax_amount: Option<String>,
    pub id: i64,
    pub business_id: i64,
}

/// `UpdateOrderTotal`: `exec`.
pub async fn update_order_total<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateOrderTotalParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE pos_orders\nSET subtotal = ?,\n    tax_amount = ?,\n    total = GREATEST(? + ? - discount_amount, 0)\nWHERE id = ? AND business_id = ?")
        .bind(&params.subtotal)
        .bind(&params.tax_amount)
        .bind(&params.subtotal)
        .bind(&params.tax_amount)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpdatePosTable` (Go `UpdatePosTableParams`).
pub struct UpdatePosTableParams {
    pub location_id: Option<i64>,
    pub name: String,
    pub capacity: Option<i32>,
    pub qr_code_url: Option<String>,
    pub business_id: i64,
    pub id: i64,
}

/// `UpdatePosTable`: `exec`.
pub async fn update_pos_table<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdatePosTableParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE pos_tables SET location_id = ?, name = ?, capacity = ?, qr_code_url = ?\nWHERE business_id = ? AND id = ?")
        .bind(params.location_id)
        .bind(&params.name)
        .bind(params.capacity)
        .bind(&params.qr_code_url)
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `UpdateSessionRefundTotals` (Go `UpdateSessionRefundTotalsParams`).
pub struct UpdateSessionRefundTotalsParams {
    pub total_refunds: Option<String>,
    pub id: i64,
}

/// `UpdateSessionRefundTotals`: `exec`.
pub async fn update_session_refund_totals<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateSessionRefundTotalsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE pos_sessions SET total_refunds = total_refunds + ? WHERE id = ?")
        .bind(&params.total_refunds)
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `UpdateSessionTotals` (Go `UpdateSessionTotalsParams`).
pub struct UpdateSessionTotalsParams {
    pub total_sales: Option<String>,
    pub total_cash_sales: Option<String>,
    pub total_card_sales: Option<String>,
    pub total_online_sales: Option<String>,
    pub total_orders: Option<i32>,
    pub id: i64,
}

/// `UpdateSessionTotals`: `exec`.
pub async fn update_session_totals<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateSessionTotalsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE pos_sessions\nSET total_sales = ?, total_cash_sales = ?, total_card_sales = ?,\n    total_online_sales = ?, total_orders = ?\nWHERE id = ?")
        .bind(&params.total_sales)
        .bind(&params.total_cash_sales)
        .bind(&params.total_card_sales)
        .bind(&params.total_online_sales)
        .bind(params.total_orders)
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `UpdateTableStatus` (Go `UpdateTableStatusParams`).
pub struct UpdateTableStatusParams {
    pub status: String,
    pub business_id: i64,
    pub id: i64,
}

/// `UpdateTableStatus`: `exec`.
pub async fn update_table_status<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateTableStatusParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE pos_tables SET status = ? WHERE business_id = ? AND id = ?")
        .bind(&params.status)
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `VoidOrder` (Go `VoidOrderParams`).
pub struct VoidOrderParams {
    pub concat: String,
    pub id: i64,
    pub business_id: i64,
}

/// `VoidOrder`: `execresult`.
pub async fn void_order<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &VoidOrderParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE pos_orders SET status = 'voided', notes = CONCAT(COALESCE(notes,''), ?)\nWHERE id = ? AND business_id = ?")
        .bind(&params.concat)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `VoidOrderItem` (Go `VoidOrderItemParams`).
pub struct VoidOrderItemParams {
    pub id: i64,
    pub order_id: i64,
}

/// `VoidOrderItem`: `execresult`.
pub async fn void_order_item<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &VoidOrderItemParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE pos_order_items SET status = 'cancelled' WHERE id = ? AND order_id = ? AND status IN ('pending','preparing')")
        .bind(params.id)
        .bind(params.order_id)
        .execute(exec)
        .await
}
