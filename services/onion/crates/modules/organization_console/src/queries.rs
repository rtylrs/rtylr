//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `AddonExists` (Go `AddonExistsParams`).
pub struct AddonExistsParams {
    pub id: String,
}

/// `AddonExists`: `one`.
pub async fn addon_exists<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &AddonExistsParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM addons WHERE id = ?")
        .bind(&params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `BanUser` (Go `BanUserParams`).
pub struct BanUserParams {
    pub ban_reason: Option<String>,
    pub id: i64,
}

/// `BanUser`: `exec`.
pub async fn ban_user<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &BanUserParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query(
        "UPDATE users\nSET is_banned = TRUE, ban_reason = ?, banned_at = NOW()\nWHERE id = ?",
    )
    .bind(&params.ban_reason)
    .bind(params.id)
    .execute(exec)
    .await
}

/// Params for `CountActiveSeatsForBusiness` (Go `CountActiveSeatsForBusinessParams`).
pub struct CountActiveSeatsForBusinessParams {
    pub business_id: i64,
}

/// `CountActiveSeatsForBusiness`: `one`.
pub async fn count_active_seats_for_business<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountActiveSeatsForBusinessParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(DISTINCT user_id)\nFROM (\n    SELECT b.owner_id AS user_id\n    FROM businesses b\n    JOIN users u ON u.id = b.owner_id\n    WHERE b.id = ? AND COALESCE(b.is_active, TRUE) = TRUE AND COALESCE(u.is_banned, FALSE) = FALSE\n    UNION\n    SELECT member.user_id AS user_id\n    FROM authorization_business_members member\n    JOIN businesses b ON b.id = member.business_id\n    JOIN users u ON u.id = member.user_id\n    WHERE member.business_id = ? AND COALESCE(b.is_active, TRUE) = TRUE\n      AND member.status = 'active' AND COALESCE(u.is_banned, FALSE) = FALSE\n) active_users")
        .bind(params.business_id)
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `CreateAddonGrant` (Go `CreateAddonGrantParams`).
pub struct CreateAddonGrantParams {
    pub business_id: i64,
    pub addon_id: String,
    pub granted_by: i64,
    pub note: Option<String>,
    pub expires_at: Option<chrono::DateTime<chrono::Utc>>,
    pub enabled_by_grant: bool,
}

/// `CreateAddonGrant`: `execresult`.
pub async fn create_addon_grant<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateAddonGrantParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO manual_addon_grants (business_id, addon_id, granted_by, note, expires_at, enabled_by_grant)\nVALUES (?, ?, ?, ?, ?, ?)")
        .bind(params.business_id)
        .bind(&params.addon_id)
        .bind(params.granted_by)
        .bind(&params.note)
        .bind(params.expires_at)
        .bind(params.enabled_by_grant)
        .execute(exec)
        .await
}

/// Params for `CreateGrant` (Go `CreateGrantParams`).
pub struct CreateGrantParams {
    pub business_id: i64,
    pub plan_id: i64,
    pub granted_by: i64,
    pub note: Option<String>,
    pub expires_at: Option<chrono::DateTime<chrono::Utc>>,
}

/// `CreateGrant`: `execresult`.
pub async fn create_grant<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateGrantParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO manual_access_grants (business_id, plan_id, granted_by, note, expires_at)\nVALUES (?, ?, ?, ?, ?)")
        .bind(params.business_id)
        .bind(params.plan_id)
        .bind(params.granted_by)
        .bind(&params.note)
        .bind(params.expires_at)
        .execute(exec)
        .await
}

/// Params for `CreateUser` (Go `CreateUserParams`).
pub struct CreateUserParams {
    pub email: String,
    pub password_hash: String,
    pub full_name: Option<String>,
    pub is_console_admin: bool,
}

/// `CreateUser`: `execresult`.
pub async fn create_user<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateUserParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO users (email, password_hash, full_name, is_console_admin)\nVALUES (?, ?, ?, ?)")
        .bind(&params.email)
        .bind(&params.password_hash)
        .bind(&params.full_name)
        .bind(params.is_console_admin)
        .execute(exec)
        .await
}

/// Params for `DeleteAddonGrant` (Go `DeleteAddonGrantParams`).
pub struct DeleteAddonGrantParams {
    pub id: i64,
    pub business_id: i64,
}

/// `DeleteAddonGrant`: `exec`.
pub async fn delete_addon_grant<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteAddonGrantParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM manual_addon_grants WHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `DeleteBusiness` (Go `DeleteBusinessParams`).
pub struct DeleteBusinessParams {
    pub id: i64,
}

/// `DeleteBusiness`: `exec`.
pub async fn delete_business<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteBusinessParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM businesses WHERE id = ?")
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `DeleteGrant` (Go `DeleteGrantParams`).
pub struct DeleteGrantParams {
    pub id: i64,
    pub business_id: i64,
}

/// `DeleteGrant`: `exec`.
pub async fn delete_grant<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteGrantParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM manual_access_grants WHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `DeleteSubscription` (Go `DeleteSubscriptionParams`).
pub struct DeleteSubscriptionParams {
    pub id: i64,
}

/// `DeleteSubscription`: `exec`.
pub async fn delete_subscription<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteSubscriptionParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM subscriptions WHERE id = ?")
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `DisableBusinessAddon` (Go `DisableBusinessAddonParams`).
pub struct DisableBusinessAddonParams {
    pub business_id: i64,
    pub addon_id: String,
}

/// `DisableBusinessAddon`: `exec`.
pub async fn disable_business_addon<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DisableBusinessAddonParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE business_addons\nSET is_enabled = FALSE, disabled_at = NOW()\nWHERE business_id = ? AND addon_id = ?")
        .bind(params.business_id)
        .bind(&params.addon_id)
        .execute(exec)
        .await
}

/// Params for `EnableBusinessAddon` (Go `EnableBusinessAddonParams`).
pub struct EnableBusinessAddonParams {
    pub business_id: i64,
    pub addon_id: String,
}

/// `EnableBusinessAddon`: `exec`.
pub async fn enable_business_addon<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &EnableBusinessAddonParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO business_addons (business_id, addon_id, is_enabled, enabled_at)\nVALUES (?, ?, TRUE, NOW())\nON DUPLICATE KEY UPDATE is_enabled = TRUE, enabled_at = NOW(), disabled_at = NULL")
        .bind(params.business_id)
        .bind(&params.addon_id)
        .execute(exec)
        .await
}

/// Params for `GetAddonGrantScope` (Go `GetAddonGrantScopeParams`).
pub struct GetAddonGrantScopeParams {
    pub id: i64,
}

/// `GetAddonGrantScope`: `one`.
pub async fn get_addon_grant_scope<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetAddonGrantScopeParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT business_id, addon_id, enabled_by_grant\nFROM manual_addon_grants\nWHERE id = ?\nORDER BY id ASC\nLIMIT 1")
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetBusinessByID` (Go `GetBusinessByIDParams`).
pub struct GetBusinessByIDParams {
    pub id: i64,
}

/// `GetBusinessByID`: `one`.
pub async fn get_business_by_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetBusinessByIDParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, name, owner_id, created_at\nFROM businesses\nWHERE id = ?\nORDER BY id ASC\nLIMIT 1")
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetGrantBusinessID` (Go `GetGrantBusinessIDParams`).
pub struct GetGrantBusinessIDParams {
    pub id: i64,
}

/// `GetGrantBusinessID`: `one`.
pub async fn get_grant_business_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetGrantBusinessIDParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT business_id FROM manual_access_grants WHERE id = ? LIMIT 1")
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetPlanByID` (Go `GetPlanByIDParams`).
pub struct GetPlanByIDParams {
    pub id: i64,
}

/// `GetPlanByID`: `one`.
pub async fn get_plan_by_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetPlanByIDParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, name, slug, price_per_seat_cents, annual_price_per_seat_cents, trial_days,\n       max_seats, max_locations, is_active, is_solo, is_featured, tier_order, tagline, created_at\nFROM plans\nWHERE id = ?\nORDER BY id ASC\nLIMIT 1")
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetPlanBySlug` (Go `GetPlanBySlugParams`).
pub struct GetPlanBySlugParams {
    pub slug: String,
}

/// `GetPlanBySlug`: `one`.
pub async fn get_plan_by_slug<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetPlanBySlugParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, name, slug, price_per_seat_cents, annual_price_per_seat_cents, trial_days,\n       max_seats, max_locations, is_active, is_solo, is_featured, tier_order, tagline, created_at\nFROM plans\nWHERE slug = ?\nORDER BY slug ASC\nLIMIT 1")
        .bind(&params.slug)
        .fetch_optional(exec)
        .await
}

/// `GetStats`: `one`.
pub async fn get_stats<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT\n    (SELECT COUNT(*) FROM users) AS total_users,\n    (SELECT COUNT(*) FROM users WHERE is_banned = TRUE) AS banned_users,\n    (SELECT COUNT(*) FROM businesses) AS total_businesses,\n    (SELECT COUNT(*) FROM subscriptions WHERE status = 'active') AS active_subs,\n    (SELECT COUNT(*) FROM subscriptions WHERE status = 'trial') AS trial_subs,\n    (SELECT COUNT(*) FROM manual_access_grants WHERE revoked_at IS NULL) AS manual_grants,\n    (SELECT COUNT(*) FROM manual_addon_grants WHERE revoked_at IS NULL) AS manual_addon_grants,\n    (SELECT COUNT(*) FROM users WHERE is_console_admin = TRUE) AS console_admins")
        .fetch_optional(exec)
        .await
}

/// Params for `GetSubscriptionByBusiness` (Go `GetSubscriptionByBusinessParams`).
pub struct GetSubscriptionByBusinessParams {
    pub business_id: i64,
}

/// `GetSubscriptionByBusiness`: `one`.
pub async fn get_subscription_by_business<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetSubscriptionByBusinessParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT s.id, s.business_id, b.name AS business_name, b.owner_id, COALESCE(u.full_name, '') AS owner_name,\n       u.email AS owner_email, s.plan_id, p.name AS plan_name,\n       s.status, s.seat_count, s.trial_ends_at, s.next_billing_date, s.current_period_ends_at,\n       s.paddle_subscription_id, s.paddle_customer_id, s.created_at, s.updated_at\nFROM subscriptions s\nJOIN plans p ON p.id = s.plan_id\nJOIN businesses b ON b.id = s.business_id\nLEFT JOIN users u ON u.id = b.owner_id\nWHERE s.business_id = ?\nLIMIT 1")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetSubscriptionByID` (Go `GetSubscriptionByIDParams`).
pub struct GetSubscriptionByIDParams {
    pub id: i64,
}

/// `GetSubscriptionByID`: `one`.
pub async fn get_subscription_by_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetSubscriptionByIDParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT s.id, s.business_id, b.name AS business_name, b.owner_id, COALESCE(u.full_name, '') AS owner_name,\n       u.email AS owner_email, s.plan_id, p.name AS plan_name,\n       s.status, s.seat_count, s.trial_ends_at, s.next_billing_date, s.current_period_ends_at,\n       s.paddle_subscription_id, s.paddle_customer_id, s.created_at, s.updated_at\nFROM subscriptions s\nJOIN plans p ON p.id = s.plan_id\nLEFT JOIN businesses b ON b.id = s.business_id\nLEFT JOIN users u ON u.id = b.owner_id\nWHERE s.id = ?\nLIMIT 1")
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
    sqlx::query("SELECT id, email, full_name, business_id, is_console_admin,\n       is_banned, ban_reason, banned_at, created_at, password_hash\nFROM users\nWHERE email = ?\nORDER BY email ASC\nLIMIT 1")
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
    sqlx::query("SELECT id, email, full_name, business_id, is_console_admin,\n       is_banned, ban_reason, banned_at, created_at, password_hash\nFROM users\nWHERE id = ?\nORDER BY id ASC\nLIMIT 1")
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `HasActiveAddonGrant` (Go `HasActiveAddonGrantParams`).
pub struct HasActiveAddonGrantParams {
    pub business_id: i64,
    pub addon_id: String,
}

/// `HasActiveAddonGrant`: `one`.
pub async fn has_active_addon_grant<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &HasActiveAddonGrantParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM manual_addon_grants\nWHERE business_id = ? AND addon_id = ? AND revoked_at IS NULL\n  AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP)")
        .bind(params.business_id)
        .bind(&params.addon_id)
        .fetch_optional(exec)
        .await
}

/// Params for `IsBusinessAddonEnabled` (Go `IsBusinessAddonEnabledParams`).
pub struct IsBusinessAddonEnabledParams {
    pub business_id: i64,
    pub addon_id: String,
}

/// `IsBusinessAddonEnabled`: `one`.
pub async fn is_business_addon_enabled<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &IsBusinessAddonEnabledParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) FROM business_addons\nWHERE business_id = ? AND addon_id = ? AND is_enabled = TRUE")
        .bind(params.business_id)
        .bind(&params.addon_id)
        .fetch_optional(exec)
        .await
}

/// Params for `ListActiveBusinessesByOwner` (Go `ListActiveBusinessesByOwnerParams`).
pub struct ListActiveBusinessesByOwnerParams {
    pub owner_id: i64,
}

/// `ListActiveBusinessesByOwner`: `many`.
pub async fn list_active_businesses_by_owner<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListActiveBusinessesByOwnerParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT b.id, b.name, b.owner_id, b.created_at\nFROM businesses b\nLEFT JOIN users u ON u.id = b.owner_id\nWHERE b.owner_id = ? AND COALESCE(b.is_active, TRUE) = TRUE\nORDER BY (b.id = u.business_id) DESC, b.id ASC\nLIMIT 100")
        .bind(params.owner_id)
        .execute(exec)
        .await
}

/// Params for `ListAddonGrants` (Go `ListAddonGrantsParams`).
pub struct ListAddonGrantsParams {
    pub limit: i32,
    pub offset: i32,
}

/// `ListAddonGrants`: `many`.
pub async fn list_addon_grants<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListAddonGrantsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT g.id, g.business_id, COALESCE(b.name, '') AS business_name,\n       b.owner_id, COALESCE(u.full_name, '') AS owner_name, u.email AS owner_email,\n       g.addon_id, a.name AS addon_name, a.price AS addon_price,\n       g.granted_by, adm.full_name AS granter_name, g.note, g.expires_at, g.revoked_at,\n       g.enabled_by_grant, g.created_at\nFROM manual_addon_grants g\nJOIN addons a ON a.id = g.addon_id\nLEFT JOIN businesses b ON b.id = g.business_id\nLEFT JOIN users u ON u.id = b.owner_id\nLEFT JOIN users adm ON adm.id = g.granted_by\nORDER BY g.created_at DESC\nLIMIT ? OFFSET ?")
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// `ListAddons`: `many`.
pub async fn list_addons<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, name, description, icon, price, is_free, sort_order\nFROM addons\nORDER BY sort_order ASC, id ASC\nLIMIT 100")
        .execute(exec)
        .await
}

/// Params for `ListGrants` (Go `ListGrantsParams`).
pub struct ListGrantsParams {
    pub limit: i32,
    pub offset: i32,
}

/// `ListGrants`: `many`.
pub async fn list_grants<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListGrantsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT g.id, g.business_id, b.name AS business_name, b.owner_id, COALESCE(u.full_name, '') AS owner_name,\n       u.email AS owner_email, g.plan_id, p.name AS plan_name,\n       g.granted_by, adm.full_name AS granter_name, g.note, g.expires_at, g.revoked_at, g.created_at\nFROM manual_access_grants g\nJOIN plans p ON p.id = g.plan_id\nLEFT JOIN businesses b ON b.id = g.business_id\nLEFT JOIN users u ON u.id = b.owner_id\nLEFT JOIN users adm ON adm.id = g.granted_by\nORDER BY g.created_at DESC\nLIMIT ? OFFSET ?")
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// `ListPlans`: `many`.
pub async fn list_plans<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, name, slug, price_per_seat_cents, annual_price_per_seat_cents, trial_days,\n       max_seats, max_locations, is_active, is_solo, is_featured, tier_order, tagline, created_at\nFROM plans\nORDER BY tier_order ASC\nLIMIT 100")
        .execute(exec)
        .await
}

/// Params for `RevokeAddonGrant` (Go `RevokeAddonGrantParams`).
pub struct RevokeAddonGrantParams {
    pub id: i64,
    pub business_id: i64,
}

/// `RevokeAddonGrant`: `exec`.
pub async fn revoke_addon_grant<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &RevokeAddonGrantParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query(
        "UPDATE manual_addon_grants SET revoked_at = NOW() WHERE id = ? AND business_id = ?",
    )
    .bind(params.id)
    .bind(params.business_id)
    .execute(exec)
    .await
}

/// Params for `RevokeGrant` (Go `RevokeGrantParams`).
pub struct RevokeGrantParams {
    pub id: i64,
    pub business_id: i64,
}

/// `RevokeGrant`: `exec`.
pub async fn revoke_grant<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &RevokeGrantParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query(
        "UPDATE manual_access_grants SET revoked_at = NOW() WHERE id = ? AND business_id = ?",
    )
    .bind(params.id)
    .bind(params.business_id)
    .execute(exec)
    .await
}

/// Params for `SetConsoleAdmin` (Go `SetConsoleAdminParams`).
pub struct SetConsoleAdminParams {
    pub is_console_admin: bool,
    pub id: i64,
}

/// `SetConsoleAdmin`: `exec`.
pub async fn set_console_admin<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &SetConsoleAdminParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE users SET is_console_admin = ? WHERE id = ?")
        .bind(params.is_console_admin)
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `UnbanUser` (Go `UnbanUserParams`).
pub struct UnbanUserParams {
    pub id: i64,
}

/// `UnbanUser`: `exec`.
pub async fn unban_user<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UnbanUserParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query(
        "UPDATE users\nSET is_banned = FALSE, ban_reason = NULL, banned_at = NULL\nWHERE id = ?",
    )
    .bind(params.id)
    .execute(exec)
    .await
}

/// Params for `UpdateBusiness` (Go `UpdateBusinessParams`).
pub struct UpdateBusinessParams {
    pub name: String,
    pub id: i64,
}

/// `UpdateBusiness`: `exec`.
pub async fn update_business<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateBusinessParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE businesses SET name = ? WHERE id = ?")
        .bind(&params.name)
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `UpdatePlan` (Go `UpdatePlanParams`).
pub struct UpdatePlanParams {
    pub name: String,
    pub price_per_seat_cents: i32,
    pub annual_price_per_seat_cents: i32,
    pub is_active: Option<bool>,
    pub id: i64,
}

/// `UpdatePlan`: `exec`.
pub async fn update_plan<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdatePlanParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE plans\nSET name = ?, price_per_seat_cents = ?, annual_price_per_seat_cents = ?, is_active = ?\nWHERE id = ?")
        .bind(&params.name)
        .bind(params.price_per_seat_cents)
        .bind(params.annual_price_per_seat_cents)
        .bind(params.is_active)
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `UpdateSubscription` (Go `UpdateSubscriptionParams`).
pub struct UpdateSubscriptionParams {
    pub plan_id: i64,
    pub status: String,
    pub seat_count: i32,
    pub trial_ends_at: Option<chrono::DateTime<chrono::Utc>>,
    pub next_billing_date: Option<chrono::DateTime<chrono::Utc>>,
    pub current_period_ends_at: Option<chrono::DateTime<chrono::Utc>>,
    pub paddle_subscription_id: Option<String>,
    pub paddle_customer_id: Option<String>,
    pub id: i64,
}

/// `UpdateSubscription`: `exec`.
pub async fn update_subscription<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateSubscriptionParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE subscriptions\nSET plan_id = ?, status = ?, seat_count = ?, trial_ends_at = ?, next_billing_date = ?,\n    current_period_ends_at = ?, paddle_subscription_id = ?, paddle_customer_id = ?,\n    updated_at = NOW()\nWHERE id = ?")
        .bind(params.plan_id)
        .bind(&params.status)
        .bind(params.seat_count)
        .bind(params.trial_ends_at)
        .bind(params.next_billing_date)
        .bind(params.current_period_ends_at)
        .bind(&params.paddle_subscription_id)
        .bind(&params.paddle_customer_id)
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `UpdateUser` (Go `UpdateUserParams`).
pub struct UpdateUserParams {
    pub full_name: Option<String>,
    pub email: String,
    pub id: i64,
}

/// `UpdateUser`: `exec`.
pub async fn update_user<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateUserParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE users\nSET full_name = ?, email = ?, updated_at = NOW()\nWHERE id = ?")
        .bind(&params.full_name)
        .bind(&params.email)
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `UpsertSubscription` (Go `UpsertSubscriptionParams`).
pub struct UpsertSubscriptionParams {
    pub business_id: i64,
    pub plan_id: i64,
    pub status: String,
    pub seat_count: i32,
    pub trial_ends_at: Option<chrono::DateTime<chrono::Utc>>,
    pub next_billing_date: Option<chrono::DateTime<chrono::Utc>>,
    pub current_period_ends_at: Option<chrono::DateTime<chrono::Utc>>,
    pub paddle_subscription_id: Option<String>,
    pub paddle_customer_id: Option<String>,
}

/// `UpsertSubscription`: `exec`.
pub async fn upsert_subscription<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpsertSubscriptionParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO subscriptions (\n    business_id, plan_id, status, seat_count, trial_ends_at, next_billing_date,\n    current_period_ends_at, paddle_subscription_id, paddle_customer_id\n) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)\nON DUPLICATE KEY UPDATE\n    plan_id = VALUES(plan_id),\n    status = VALUES(status),\n    seat_count = VALUES(seat_count),\n    trial_ends_at = VALUES(trial_ends_at),\n    next_billing_date = VALUES(next_billing_date),\n    current_period_ends_at = VALUES(current_period_ends_at),\n    paddle_subscription_id = VALUES(paddle_subscription_id),\n    paddle_customer_id = VALUES(paddle_customer_id),\n    updated_at = NOW()")
        .bind(params.business_id)
        .bind(params.plan_id)
        .bind(&params.status)
        .bind(params.seat_count)
        .bind(params.trial_ends_at)
        .bind(params.next_billing_date)
        .bind(params.current_period_ends_at)
        .bind(&params.paddle_subscription_id)
        .bind(&params.paddle_customer_id)
        .execute(exec)
        .await
}
