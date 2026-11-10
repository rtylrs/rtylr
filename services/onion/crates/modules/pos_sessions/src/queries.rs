//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CloseSession` (Go `CloseSessionParams`).
pub struct CloseSessionParams {
    pub closing_cash: Option<String>,
    pub id: i64,
    pub business_id: i64,
}

/// `CloseSession`: `execresult`.
pub async fn close_session<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CloseSessionParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE pos_sessions SET closing_cash = ?, status = 'closed', closed_at = NOW()\nWHERE id = ? AND business_id = ? AND status = 'open'")
        .bind(&params.closing_cash)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `GetSessionByID` (Go `GetSessionByIDParams`).
pub struct GetSessionByIDParams {
    pub id: i64,
    pub business_id: i64,
}

/// `GetSessionByID`: `one`.
pub async fn get_session_by_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetSessionByIDParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, location_id, staff_id, status, opening_cash, closing_cash, expected_cash,\n       cash_difference, total_sales, total_orders, total_cash_sales, total_card_sales,\n       total_online_sales, total_refunds, opened_at, closed_at, notes\nFROM pos_sessions WHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `ListSessionsByOwner` (Go `ListSessionsByOwnerParams`).
pub struct ListSessionsByOwnerParams {
    pub business_id: i64,
}

/// `ListSessionsByOwner`: `many`.
pub async fn list_sessions_by_owner<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListSessionsByOwnerParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, location_id, staff_id, status, opening_cash, closing_cash, expected_cash,\n       cash_difference, total_sales, total_orders, total_cash_sales, total_card_sales,\n       total_online_sales, total_refunds, opened_at, closed_at, notes\nFROM pos_sessions WHERE business_id = ? ORDER BY opened_at DESC LIMIT 100")
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
    sqlx::query("INSERT INTO pos_sessions (owner_id, business_id, location_id, staff_id, opening_cash, status, opened_at)\nVALUES (?, ?, ?, ?, ?, 'open', NOW())")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.location_id)
        .bind(params.staff_id)
        .bind(&params.opening_cash)
        .execute(exec)
        .await
}
