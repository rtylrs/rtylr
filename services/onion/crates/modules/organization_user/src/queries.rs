//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `BusinessNameExists` (Go `BusinessNameExistsParams`).
pub struct BusinessNameExistsParams {
    pub name: String,
}

/// `BusinessNameExists`: `one`.
pub async fn business_name_exists<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &BusinessNameExistsParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT EXISTS(SELECT 1 FROM businesses WHERE name = ? LIMIT 1) AS exists_flag")
        .bind(&params.name)
        .fetch_optional(exec)
        .await
}

/// Params for `CheckBusinessAccess` (Go `CheckBusinessAccessParams`).
pub struct CheckBusinessAccessParams {
    pub id: i64,
    pub owner_id: i64,
    pub user_id: i64,
    pub business_id: i64,
}

/// `CheckBusinessAccess`: `one`.
pub async fn check_business_access<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CheckBusinessAccessParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) AS cnt FROM (\n    SELECT 1 FROM businesses b WHERE b.id = ? AND b.owner_id = ? AND b.is_active = TRUE\n    UNION ALL\n    SELECT 1 FROM authorization_business_members member\n    JOIN businesses b ON b.id = member.business_id AND b.is_active = TRUE\n    WHERE member.user_id = ? AND member.business_id = ? AND member.status = 'active'\n) x")
        .bind(params.id)
        .bind(params.owner_id)
        .bind(params.user_id)
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CreateBusiness` (Go `CreateBusinessParams`).
pub struct CreateBusinessParams {
    pub name: String,
    pub owner_id: i64,
}

/// `CreateBusiness`: `execresult`.
pub async fn create_business<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateBusinessParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO businesses (name, owner_id) VALUES (?, ?)")
        .bind(&params.name)
        .bind(params.owner_id)
        .execute(exec)
        .await
}

/// Params for `CreateUser` (Go `CreateUserParams`).
pub struct CreateUserParams {
    pub email: String,
    pub password_hash: String,
    pub full_name: Option<String>,
    pub signup_ip_address: Option<String>,
    pub signup_device_info: Option<String>,
}

/// `CreateUser`: `execresult`.
pub async fn create_user<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateUserParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO users (\n    email, password_hash, full_name, signup_ip_address, signup_device_info\n) VALUES (\n    ?, ?, ?, ?, ?\n)")
        .bind(&params.email)
        .bind(&params.password_hash)
        .bind(&params.full_name)
        .bind(&params.signup_ip_address)
        .bind(&params.signup_device_info)
        .execute(exec)
        .await
}

/// Params for `DeleteUser` (Go `DeleteUserParams`).
pub struct DeleteUserParams {
    pub id: i64,
}

/// `DeleteUser`: `exec`.
pub async fn delete_user<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteUserParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM users WHERE id = ?")
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `EmailExists` (Go `EmailExistsParams`).
pub struct EmailExistsParams {
    pub email: String,
}

/// `EmailExists`: `one`.
pub async fn email_exists<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &EmailExistsParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)")
        .bind(&params.email)
        .fetch_optional(exec)
        .await
}

/// Params for `GetActivePlanForSignup` (Go `GetActivePlanForSignupParams`).
pub struct GetActivePlanForSignupParams {
    pub slug: String,
}

/// `GetActivePlanForSignup`: `one`.
pub async fn get_active_plan_for_signup<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetActivePlanForSignupParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, trial_days FROM plans WHERE slug = ? AND is_active = TRUE")
        .bind(&params.slug)
        .fetch_optional(exec)
        .await
}

/// Params for `GetBusinessOwnerID` (Go `GetBusinessOwnerIDParams`).
pub struct GetBusinessOwnerIDParams {
    pub id: i64,
}

/// `GetBusinessOwnerID`: `one`.
pub async fn get_business_owner_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetBusinessOwnerIDParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT owner_id FROM businesses WHERE id = ? AND is_active = TRUE LIMIT 1")
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetOwnedDefaultBusinessID` (Go `GetOwnedDefaultBusinessIDParams`).
pub struct GetOwnedDefaultBusinessIDParams {
    pub id: i64,
}

/// `GetOwnedDefaultBusinessID`: `one`.
pub async fn get_owned_default_business_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetOwnedDefaultBusinessIDParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT b.id\nFROM users u\nJOIN businesses b ON b.id = u.business_id AND b.owner_id = u.id\nWHERE u.id = ?")
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetUserAccessState` (Go `GetUserAccessStateParams`).
pub struct GetUserAccessStateParams {
    pub id: i64,
}

/// `GetUserAccessState`: `one`.
pub async fn get_user_access_state<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetUserAccessStateParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT\n    EXISTS(\n        SELECT 1 FROM authorization_business_members member\n        WHERE member.user_id = u.id\n    ) AS has_permissions,\n    EXISTS(\n        SELECT 1 FROM authorization_business_members member\n        WHERE member.user_id = u.id AND member.status = 'active'\n    ) OR EXISTS(\n        SELECT 1 FROM businesses b\n        WHERE b.owner_id = u.id AND COALESCE(b.is_active, TRUE) = TRUE\n    ) AS has_active_access\nFROM users u\nWHERE u.id = ?")
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetUserBusinessID` (Go `GetUserBusinessIDParams`).
pub struct GetUserBusinessIDParams {
    pub id: i64,
}

/// `GetUserBusinessID`: `one`.
pub async fn get_user_business_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetUserBusinessIDParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT business_id FROM users WHERE id = ?")
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetUserByEmail` (Go `GetUserByEmailParams`).
pub struct GetUserByEmailParams {
    pub email: String,
}

/// `GetUserByEmail`: `one`.
pub async fn get_user_by_email<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetUserByEmailParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, business_id, email, password_hash, full_name, signup_ip_address, signup_device_info, created_at, updated_at, is_console_admin, is_banned, token_epoch\nFROM users WHERE email = ?")
        .bind(&params.email)
        .fetch_optional(exec)
        .await
}

/// Params for `GetUserByID` (Go `GetUserByIDParams`).
pub struct GetUserByIDParams {
    pub id: i64,
}

/// `GetUserByID`: `one`.
pub async fn get_user_by_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetUserByIDParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, business_id, email, password_hash, full_name, signup_ip_address, signup_device_info, created_at, updated_at, is_console_admin, is_banned, token_epoch\nFROM users WHERE id = ?")
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetUserByIdentifier` (Go `GetUserByIdentifierParams`).
pub struct GetUserByIdentifierParams {
    pub email: String,
    pub lower: String,
}

/// `GetUserByIdentifier`: `one`.
pub async fn get_user_by_identifier<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetUserByIdentifierParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, email, password_hash, full_name, signup_ip_address, signup_device_info, created_at, updated_at, is_console_admin, is_banned, token_epoch\nFROM users\nWHERE email = ? OR LOWER(username) = LOWER(?)\nORDER BY id ASC\nLIMIT 1")
        .bind(&params.email)
        .bind(&params.lower)
        .fetch_optional(exec)
        .await
}

/// Params for `RecordLoginHistory` (Go `RecordLoginHistoryParams`).
pub struct RecordLoginHistoryParams {
    pub user_id: i64,
    pub email: String,
    pub method: String,
    pub ip_address: Option<String>,
    pub user_agent: Option<String>,
    pub success: Option<bool>,
}

/// `RecordLoginHistory`: `exec`.
pub async fn record_login_history<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &RecordLoginHistoryParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO login_history (user_id, email, method, ip_address, user_agent, success) VALUES (?, ?, ?, ?, ?, ?)")
        .bind(params.user_id)
        .bind(&params.email)
        .bind(&params.method)
        .bind(&params.ip_address)
        .bind(&params.user_agent)
        .bind(params.success)
        .execute(exec)
        .await
}

/// Params for `SetUserBusinessID` (Go `SetUserBusinessIDParams`).
pub struct SetUserBusinessIDParams {
    pub business_id: Option<i64>,
    pub id: i64,
}

/// `SetUserBusinessID`: `exec`.
pub async fn set_user_business_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &SetUserBusinessIDParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE users SET business_id = ? WHERE id = ?")
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `UpdateUserPassword` (Go `UpdateUserPasswordParams`).
pub struct UpdateUserPasswordParams {
    pub password_hash: String,
    pub id: i64,
}

/// `UpdateUserPassword`: `exec`.
pub async fn update_user_password<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateUserPasswordParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE users\nSET password_hash = ?, token_epoch = token_epoch + 1, updated_at = CURRENT_TIMESTAMP\nWHERE id = ?")
        .bind(&params.password_hash)
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `UpdateUserProfile` (Go `UpdateUserProfileParams`).
pub struct UpdateUserProfileParams {
    pub full_name: Option<String>,
    pub id: i64,
}

/// `UpdateUserProfile`: `exec`.
pub async fn update_user_profile<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateUserProfileParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE users\nSET full_name = ?, updated_at = CURRENT_TIMESTAMP\nWHERE id = ?")
        .bind(&params.full_name)
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `UpsertSignupSubscription` (Go `UpsertSignupSubscriptionParams`).
pub struct UpsertSignupSubscriptionParams {
    pub business_id: i64,
    pub plan_id: i64,
    pub status: String,
    pub trial_ends_at: Option<chrono::DateTime<chrono::Utc>>,
}

/// `UpsertSignupSubscription`: `exec`.
pub async fn upsert_signup_subscription<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpsertSignupSubscriptionParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO subscriptions (business_id, plan_id, status, trial_ends_at)\nVALUES (?, ?, ?, ?)\nON DUPLICATE KEY UPDATE\n  plan_id = VALUES(plan_id),\n  status = VALUES(status),\n  trial_ends_at = VALUES(trial_ends_at),\n  updated_at = CURRENT_TIMESTAMP")
        .bind(params.business_id)
        .bind(params.plan_id)
        .bind(&params.status)
        .bind(params.trial_ends_at)
        .execute(exec)
        .await
}
