//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CountTransactions` (Go `CountTransactionsParams`).
pub struct CountTransactionsParams {
    pub business_id: i64,
}

/// `CountTransactions`: `one`.
pub async fn count_transactions<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountTransactionsParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM crm_loyalty_transactions WHERE business_id = ?")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CountTransactionsByCustomer` (Go `CountTransactionsByCustomerParams`).
pub struct CountTransactionsByCustomerParams {
    pub business_id: i64,
    pub customer_id: i64,
}

/// `CountTransactionsByCustomer`: `one`.
pub async fn count_transactions_by_customer<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountTransactionsByCustomerParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query(
        "SELECT COUNT(*) FROM crm_loyalty_transactions WHERE business_id = ? AND customer_id = ?",
    )
    .bind(params.business_id)
    .bind(params.customer_id)
    .fetch_optional(exec)
    .await
}

/// Params for `CreateRedeemTransaction` (Go `CreateRedeemTransactionParams`).
pub struct CreateRedeemTransactionParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub customer_id: i64,
    pub points: i32,
    pub description: Option<String>,
    pub reference_id: Option<i64>,
    pub reference_type: Option<String>,
}

/// `CreateRedeemTransaction`: `exec`.
pub async fn create_redeem_transaction<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateRedeemTransactionParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO crm_loyalty_transactions (owner_id, business_id, customer_id, points, type, description, reference_id, reference_type)\nVALUES (?, ?, ?, ?, 'redeem', ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.customer_id)
        .bind(params.points)
        .bind(&params.description)
        .bind(params.reference_id)
        .bind(&params.reference_type)
        .execute(exec)
        .await
}

/// Params for `CreateReward` (Go `CreateRewardParams`).
pub struct CreateRewardParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub name: String,
    pub description: Option<String>,
    pub category: Option<String>,
    pub points_cost: i32,
    pub stock: Option<i32>,
    pub icon: Option<String>,
    pub active: Option<bool>,
}

/// `CreateReward`: `execresult`.
pub async fn create_reward<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateRewardParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO crm_loyalty_rewards (owner_id, business_id, name, description, category, points_cost, stock, icon, active)\nVALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(&params.name)
        .bind(&params.description)
        .bind(&params.category)
        .bind(params.points_cost)
        .bind(params.stock)
        .bind(&params.icon)
        .bind(params.active)
        .execute(exec)
        .await
}

/// Params for `CreateTransaction` (Go `CreateTransactionParams`).
pub struct CreateTransactionParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub customer_id: i64,
    pub points: i32,
    pub r#type: String,
    pub description: Option<String>,
    pub reference_id: Option<i64>,
    pub reference_type: Option<String>,
}

/// `CreateTransaction`: `exec`.
pub async fn create_transaction<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateTransactionParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO crm_loyalty_transactions (owner_id, business_id, customer_id, points, type, description, reference_id, reference_type)\nVALUES (?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.customer_id)
        .bind(params.points)
        .bind(&params.r#type)
        .bind(&params.description)
        .bind(params.reference_id)
        .bind(&params.reference_type)
        .execute(exec)
        .await
}

/// Params for `GetConfig` (Go `GetConfigParams`).
pub struct GetConfigParams {
    pub business_id: i64,
}

/// `GetConfig`: `one`.
pub async fn get_config<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetConfigParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, points_per_currency_unit, currency_per_point, min_redemption_points, is_enabled, updated_at\nFROM crm_loyalty_config WHERE business_id = ?")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetCustomerPointsBalance` (Go `GetCustomerPointsBalanceParams`).
pub struct GetCustomerPointsBalanceParams {
    pub business_id: i64,
    pub customer_id: i64,
}

/// `GetCustomerPointsBalance`: `one`.
pub async fn get_customer_points_balance<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetCustomerPointsBalanceParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT CAST(COALESCE(SUM(points), 0) AS SIGNED) AS balance\nFROM crm_loyalty_transactions WHERE business_id = ? AND customer_id = ?")
        .bind(params.business_id)
        .bind(params.customer_id)
        .fetch_optional(exec)
        .await
}

/// Params for `ListActiveLoyaltyRewards` (Go `ListActiveLoyaltyRewardsParams`).
pub struct ListActiveLoyaltyRewardsParams {
    pub business_id: i64,
}

/// `ListActiveLoyaltyRewards`: `many`.
pub async fn list_active_loyalty_rewards<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListActiveLoyaltyRewardsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, name, COALESCE(description, '') AS description,\n       COALESCE(category, 'general') AS category, points_cost, stock,\n       COALESCE(icon, '') AS icon, COALESCE(active, TRUE) AS active,\n       created_at, updated_at\nFROM crm_loyalty_rewards\nWHERE business_id = ? AND active = TRUE\nORDER BY points_cost ASC, id ASC\nLIMIT 200")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `ListLoyaltyMembers` (Go `ListLoyaltyMembersParams`).
pub struct ListLoyaltyMembersParams {
    pub business_id: i64,
    pub limit: i32,
    pub offset: i32,
}

/// `ListLoyaltyMembers`: `many`.
pub async fn list_loyalty_members<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListLoyaltyMembersParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT\n    id,\n    owner_id,\n    CONCAT_WS(' ', first_name, COALESCE(last_name, '')) AS name,\n    COALESCE(loyalty_points, 0) AS loyalty_points,\n    COALESCE(loyalty_tier, 'bronze') AS loyalty_tier,\n    COALESCE(visit_count, 0) AS visit_count,\n    created_at,\n    COALESCE(last_visit_at, created_at) AS last_active_at\nFROM crm_customers\nWHERE business_id = ? AND is_active = TRUE\n  AND (loyalty_points > 0 OR visit_count > 0)\nORDER BY loyalty_points DESC, visit_count DESC\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `ListLoyaltyRewards` (Go `ListLoyaltyRewardsParams`).
pub struct ListLoyaltyRewardsParams {
    pub business_id: i64,
}

/// `ListLoyaltyRewards`: `many`.
pub async fn list_loyalty_rewards<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListLoyaltyRewardsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, name, COALESCE(description, '') AS description,\n       COALESCE(category, 'general') AS category, points_cost, stock,\n       COALESCE(icon, '') AS icon, COALESCE(active, TRUE) AS active,\n       created_at, updated_at\nFROM crm_loyalty_rewards\nWHERE business_id = ?\nORDER BY points_cost ASC, id ASC\nLIMIT 200")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `ListTransactions` (Go `ListTransactionsParams`).
pub struct ListTransactionsParams {
    pub business_id: i64,
    pub limit: i32,
    pub offset: i32,
}

/// `ListTransactions`: `many`.
pub async fn list_transactions<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListTransactionsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, customer_id, points, type, description, reference_id, reference_type, created_at\nFROM crm_loyalty_transactions WHERE business_id = ?\nORDER BY created_at DESC LIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `ListTransactionsByCustomer` (Go `ListTransactionsByCustomerParams`).
pub struct ListTransactionsByCustomerParams {
    pub business_id: i64,
    pub customer_id: i64,
    pub limit: i32,
    pub offset: i32,
}

/// `ListTransactionsByCustomer`: `many`.
pub async fn list_transactions_by_customer<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListTransactionsByCustomerParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, customer_id, points, type, description, reference_id, reference_type, created_at\nFROM crm_loyalty_transactions WHERE business_id = ? AND customer_id = ?\nORDER BY created_at DESC LIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(params.customer_id)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `UpdateReward` (Go `UpdateRewardParams`).
pub struct UpdateRewardParams {
    pub name: String,
    pub description: Option<String>,
    pub category: Option<String>,
    pub points_cost: i32,
    pub stock: Option<i32>,
    pub icon: Option<String>,
    pub active: Option<bool>,
    pub id: i64,
    pub business_id: i64,
}

/// `UpdateReward`: `execresult`.
pub async fn update_reward<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateRewardParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE crm_loyalty_rewards\nSET name = ?, description = ?, category = ?, points_cost = ?, stock = ?, icon = ?, active = ?\nWHERE id = ? AND business_id = ?")
        .bind(&params.name)
        .bind(&params.description)
        .bind(&params.category)
        .bind(params.points_cost)
        .bind(params.stock)
        .bind(&params.icon)
        .bind(params.active)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpsertConfig` (Go `UpsertConfigParams`).
pub struct UpsertConfigParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub points_per_currency_unit: Option<String>,
    pub currency_per_point: Option<String>,
    pub min_redemption_points: Option<i32>,
    pub is_enabled: Option<bool>,
}

/// `UpsertConfig`: `exec`.
pub async fn upsert_config<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpsertConfigParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO crm_loyalty_config (owner_id, business_id, points_per_currency_unit, currency_per_point, min_redemption_points, is_enabled)\nVALUES (?, ?, ?, ?, ?, ?)\nON DUPLICATE KEY UPDATE\n  points_per_currency_unit = VALUES(points_per_currency_unit),\n  currency_per_point = VALUES(currency_per_point),\n  min_redemption_points = VALUES(min_redemption_points),\n  is_enabled = VALUES(is_enabled)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(&params.points_per_currency_unit)
        .bind(&params.currency_per_point)
        .bind(params.min_redemption_points)
        .bind(params.is_enabled)
        .execute(exec)
        .await
}
