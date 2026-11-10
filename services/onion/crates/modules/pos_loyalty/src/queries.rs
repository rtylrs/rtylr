//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CreateLoyaltyTransaction` (Go `CreateLoyaltyTransactionParams`).
pub struct CreateLoyaltyTransactionParams {
    pub business_id: i64,
    pub customer_id: i64,
    pub order_id: Option<i64>,
    pub r#type: String,
    pub points: i32,
    pub balance_after: i32,
    pub note: Option<String>,
}

/// `CreateLoyaltyTransaction`: `execresult`.
pub async fn create_loyalty_transaction<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateLoyaltyTransactionParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO pos_loyalty_transactions (business_id, customer_id, order_id, type, points, balance_after, note)\nVALUES (?, ?, ?, ?, ?, ?, ?)")
        .bind(params.business_id)
        .bind(params.customer_id)
        .bind(params.order_id)
        .bind(&params.r#type)
        .bind(params.points)
        .bind(params.balance_after)
        .bind(&params.note)
        .execute(exec)
        .await
}

/// Params for `GetLoyaltyBalance` (Go `GetLoyaltyBalanceParams`).
pub struct GetLoyaltyBalanceParams {
    pub business_id: i64,
    pub customer_id: i64,
}

/// `GetLoyaltyBalance`: `one`.
pub async fn get_loyalty_balance<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetLoyaltyBalanceParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COALESCE(\n  (SELECT balance_after FROM pos_loyalty_transactions\n   WHERE business_id = ? AND customer_id = ?\n   ORDER BY id DESC LIMIT 1),\n  0\n) AS balance")
        .bind(params.business_id)
        .bind(params.customer_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetLoyaltyConfig` (Go `GetLoyaltyConfigParams`).
pub struct GetLoyaltyConfigParams {
    pub business_id: i64,
}

/// `GetLoyaltyConfig`: `one`.
pub async fn get_loyalty_config<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetLoyaltyConfigParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, business_id, is_enabled, points_per_dollar, redemption_rate, min_redeem_points,\n       created_at, updated_at\nFROM pos_loyalty_config WHERE business_id = ?")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `ListLoyaltyTransactions` (Go `ListLoyaltyTransactionsParams`).
pub struct ListLoyaltyTransactionsParams {
    pub business_id: i64,
    pub customer_id: i64,
}

/// `ListLoyaltyTransactions`: `many`.
pub async fn list_loyalty_transactions<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListLoyaltyTransactionsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, business_id, customer_id, order_id, type, points, balance_after, note, created_at\nFROM pos_loyalty_transactions\nWHERE business_id = ? AND customer_id = ?\nORDER BY created_at DESC LIMIT 100")
        .bind(params.business_id)
        .bind(params.customer_id)
        .execute(exec)
        .await
}

/// Params for `UpsertLoyaltyConfig` (Go `UpsertLoyaltyConfigParams`).
pub struct UpsertLoyaltyConfigParams {
    pub business_id: i64,
    pub is_enabled: Option<bool>,
    pub points_per_dollar: Option<String>,
    pub redemption_rate: Option<String>,
    pub min_redeem_points: Option<i32>,
}

/// `UpsertLoyaltyConfig`: `execresult`.
pub async fn upsert_loyalty_config<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpsertLoyaltyConfigParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO pos_loyalty_config (business_id, is_enabled, points_per_dollar, redemption_rate, min_redeem_points)\nVALUES (?, ?, ?, ?, ?)\nON DUPLICATE KEY UPDATE\n    is_enabled = VALUES(is_enabled),\n    points_per_dollar = VALUES(points_per_dollar),\n    redemption_rate = VALUES(redemption_rate),\n    min_redeem_points = VALUES(min_redeem_points)")
        .bind(params.business_id)
        .bind(params.is_enabled)
        .bind(&params.points_per_dollar)
        .bind(&params.redemption_rate)
        .bind(params.min_redeem_points)
        .execute(exec)
        .await
}
