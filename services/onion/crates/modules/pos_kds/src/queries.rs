//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `GetKDSOrderItemsForOrders` (Go `GetKDSOrderItemsForOrdersParams`).
pub struct GetKDSOrderItemsForOrdersParams {
    pub order_ids: Vec<i64>,
}

/// `GetKDSOrderItemsForOrders`: `many`.
pub async fn get_kds_order_items_for_orders<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetKDSOrderItemsForOrdersParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    let mut sql_template: String = "SELECT id, order_id, name, quantity, notes, status\nFROM pos_order_items\nWHERE order_id IN ({SLICE:order_ids})\nORDER BY id".to_string();

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

/// Params for `ListActiveKDSOrders` (Go `ListActiveKDSOrdersParams`).
pub struct ListActiveKDSOrdersParams {
    pub business_id: i64,
}

/// `ListActiveKDSOrders`: `many`.
pub async fn list_active_kds_orders<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListActiveKDSOrdersParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, order_number, type, status, table_id, customer_name, kitchen_notes, created_at, owner_id, business_id\nFROM pos_orders\nWHERE business_id = ? AND status IN ('confirmed','preparing','ready')\nORDER BY created_at ASC")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpdateKDSItemStatusByItemID` (Go `UpdateKDSItemStatusByItemIDParams`).
pub struct UpdateKDSItemStatusByItemIDParams {
    pub status: String,
    pub id: i64,
    pub business_id: i64,
}

/// `UpdateKDSItemStatusByItemID`: `exec`.
pub async fn update_kds_item_status_by_item_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateKDSItemStatusByItemIDParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE pos_order_items oi\nJOIN pos_orders o ON o.id = oi.order_id\nSET oi.status = ?\nWHERE oi.id = ? AND o.business_id = ?")
        .bind(&params.status)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}
