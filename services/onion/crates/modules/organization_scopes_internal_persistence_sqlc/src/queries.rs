//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `AccountByEmail` (Go `AccountByEmailParams`).
pub struct AccountByEmailParams {
    pub email: String,
}

/// `AccountByEmail`: `one`.
pub async fn account_by_email<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &AccountByEmailParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id FROM users WHERE email = ? LIMIT 1")
        .bind(&params.email)
        .fetch_optional(exec)
        .await
}

/// Params for `AccountEmail` (Go `AccountEmailParams`).
pub struct AccountEmailParams {
    pub id: i64,
}

/// `AccountEmail`: `one`.
pub async fn account_email<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &AccountEmailParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT email FROM users WHERE id = ?")
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `AccountPasswordHash` (Go `AccountPasswordHashParams`).
pub struct AccountPasswordHashParams {
    pub id: i64,
}

/// `AccountPasswordHash`: `one`.
pub async fn account_password_hash<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &AccountPasswordHashParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT password_hash FROM users WHERE id = ?")
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `ActiveBusinessUsers` (Go `ActiveBusinessUsersParams`).
pub struct ActiveBusinessUsersParams {
    pub business_id: i64,
}

/// `ActiveBusinessUsers`: `many`.
pub async fn active_business_users<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ActiveBusinessUsersParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT u.id, COALESCE(u.full_name, '') AS name, u.email\nFROM users u\nJOIN businesses b ON b.id = ? AND COALESCE(b.is_active, TRUE) = TRUE\nWHERE b.owner_id = u.id OR EXISTS (\n  SELECT 1 FROM authorization_business_members m\n  WHERE m.business_id = b.id AND m.user_id = u.id AND m.status = 'active')\nORDER BY name, u.id")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `AssistantIdentity` (Go `AssistantIdentityParams`).
pub struct AssistantIdentityParams {
    pub business_id: i64,
    pub user_id: i64,
}

/// `AssistantIdentity`: `one`.
pub async fn assistant_identity<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &AssistantIdentityParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COALESCE(u.full_name, '') AS name, b.name AS business, COALESCE(b.timezone,'UTC') AS timezone, COALESCE(b.currency,'USD') AS currency\nFROM users u JOIN businesses b ON b.id = ?\nWHERE u.id = ? AND b.is_active = TRUE\n AND (b.owner_id = u.id OR EXISTS (SELECT 1 FROM authorization_business_members m\n WHERE m.business_id=b.id AND m.user_id=u.id AND m.status='active'))")
        .bind(params.business_id)
        .bind(params.user_id)
        .fetch_optional(exec)
        .await
}

/// Params for `AuthorizationVersion` (Go `AuthorizationVersionParams`).
pub struct AuthorizationVersionParams {
    pub user_id: i64,
    pub business_id: i64,
}

/// `AuthorizationVersion`: `one`.
pub async fn authorization_version<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &AuthorizationVersionParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COALESCE(m.authorization_version, 1) AS version\nFROM businesses b\nLEFT JOIN authorization_business_members m ON m.business_id = b.id AND m.user_id = ? AND m.status = 'active'\nWHERE b.id = ? AND b.is_active = TRUE\n AND (b.owner_id = ? OR m.user_id IS NOT NULL)")
        .bind(params.user_id)
        .bind(params.business_id)
        .bind(params.user_id)
        .fetch_optional(exec)
        .await
}

/// Params for `BusinessUserByUsername` (Go `BusinessUserByUsernameParams`).
pub struct BusinessUserByUsernameParams {
    pub username: Option<String>,
    pub business_id: i64,
}

/// `BusinessUserByUsername`: `one`.
pub async fn business_user_by_username<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &BusinessUserByUsernameParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT u.id FROM users u WHERE u.username = ? AND EXISTS (\n SELECT 1 FROM businesses b WHERE b.id = ? AND b.is_active = TRUE\n AND (b.owner_id = u.id OR EXISTS (\n SELECT 1 FROM authorization_business_members m WHERE m.business_id = b.id AND m.user_id = u.id AND m.status = 'active'))\n) LIMIT 1")
        .bind(&params.username)
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `BusinessUserName` (Go `BusinessUserNameParams`).
pub struct BusinessUserNameParams {
    pub user_id: i64,
    pub business_id: i64,
}

/// `BusinessUserName`: `one`.
pub async fn business_user_name<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &BusinessUserNameParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT u.full_name FROM users u\nWHERE u.id = ? AND EXISTS (\n SELECT 1 FROM businesses b WHERE b.id = ?\n AND (b.owner_id = u.id OR EXISTS (SELECT 1 FROM authorization_business_members m WHERE m.business_id = b.id AND m.user_id = u.id))\n)")
        .bind(params.user_id)
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `BusinessUsers` (Go `BusinessUsersParams`).
pub struct BusinessUsersParams {
    pub user_ids: Vec<i64>,
    pub business_id: i64,
}

/// `BusinessUsers`: `many`.
pub async fn business_users<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &BusinessUsersParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    let mut sql_template: String = "SELECT u.id, COALESCE(u.full_name, '') AS name, u.email\nFROM users u\nWHERE u.id IN ({SLICE:user_ids}) AND EXISTS (\n SELECT 1 FROM businesses b WHERE b.id = ?\n AND (b.owner_id = u.id OR u.business_id = b.id OR EXISTS (\n SELECT 1 FROM authorization_business_members m WHERE m.business_id = b.id AND m.user_id = u.id))\n)".to_string();

    if params.user_ids.is_empty() {
        sql_template = sql_template.replace("{SLICE:user_ids}", "NULL");
    } else {
        sql_template = sql_template.replace(
            "{SLICE:user_ids}",
            &std::iter::repeat_n("?", params.user_ids.len())
                .collect::<Vec<_>>()
                .join(","),
        );
    }
    let query = sqlx::query(&sql_template);
    let mut query = query.bind(params.business_id);
    for value in &params.user_ids {
        query = query.bind(value);
    }
    query.execute(exec).await
}

/// Params for `GetActiveBusinessOwnerID` (Go `GetActiveBusinessOwnerIDParams`).
pub struct GetActiveBusinessOwnerIDParams {
    pub id: i64,
}

/// `GetActiveBusinessOwnerID`: `one`.
pub async fn get_active_business_owner_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetActiveBusinessOwnerIDParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT owner_id FROM businesses WHERE id = ? AND is_active = TRUE LIMIT 1")
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetAuthorizedBusinessID` (Go `GetAuthorizedBusinessIDParams`).
pub struct GetAuthorizedBusinessIDParams {
    pub id: i64,
    pub user_id: i64,
}

/// `GetAuthorizedBusinessID`: `one`.
pub async fn get_authorized_business_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetAuthorizedBusinessIDParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT b.id\nFROM businesses b\nWHERE b.id = ? AND COALESCE(b.is_active, TRUE) = TRUE\n  AND (b.owner_id = ? OR EXISTS (\n    SELECT 1 FROM authorization_business_members abm\n    WHERE abm.user_id = ? AND abm.business_id = b.id AND abm.status = 'active'\n  ))\nLIMIT 1")
        .bind(params.id)
        .bind(params.user_id)
        .bind(params.user_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetUserDefaultBusinessID` (Go `GetUserDefaultBusinessIDParams`).
pub struct GetUserDefaultBusinessIDParams {
    pub id: i64,
}

/// `GetUserDefaultBusinessID`: `one`.
pub async fn get_user_default_business_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetUserDefaultBusinessIDParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT business_id FROM users WHERE id = ? LIMIT 1")
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `LocationLimit` (Go `LocationLimitParams`).
pub struct LocationLimitParams {
    pub id: i64,
}

/// `LocationLimit`: `one`.
pub async fn location_limit<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &LocationLimitParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT p.max_locations FROM businesses b\nJOIN subscriptions s ON s.business_id = b.id\nJOIN plans p ON p.id = s.plan_id\nWHERE b.id = ? AND b.is_active = TRUE")
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `SubscriptionAccess` (Go `SubscriptionAccessParams`).
pub struct SubscriptionAccessParams {
    pub business_id: i64,
}

/// `SubscriptionAccess`: `one`.
pub async fn subscription_access<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &SubscriptionAccessParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT s.status,\n       EXISTS (SELECT 1 FROM manual_access_grants g\n           WHERE g.business_id = s.business_id AND g.plan_id = s.plan_id\n             AND g.revoked_at IS NULL\n             AND (g.expires_at IS NULL OR g.expires_at > CURRENT_TIMESTAMP)) AS granted,\n       s.trial_ends_at\nFROM subscriptions s WHERE s.business_id = ?")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `UploadKeyScope` (Go `UploadKeyScopeParams`).
pub struct UploadKeyScopeParams {
    pub key_hash: String,
}

/// `UploadKeyScope`: `one`.
pub async fn upload_key_scope<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UploadKeyScopeParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT owner_id, business_id FROM hub_api_keys\nWHERE key_hash = ? AND is_active = TRUE\n  AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP)\n  AND JSON_CONTAINS(scopes, JSON_QUOTE('upload'))\nLIMIT 1")
        .bind(&params.key_hash)
        .fetch_optional(exec)
        .await
}
