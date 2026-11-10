//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CloseActiveSeatPeriod` (Go `CloseActiveSeatPeriodParams`).
pub struct CloseActiveSeatPeriodParams {
    pub period_end: Option<chrono::DateTime<chrono::Utc>>,
    pub subscription_id: i64,
}

/// `CloseActiveSeatPeriod`: `exec`.
pub async fn close_active_seat_period<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CloseActiveSeatPeriodParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query(
        "UPDATE seat_periods SET period_end = ?\nWHERE subscription_id = ? AND period_end IS NULL",
    )
    .bind(params.period_end)
    .bind(params.subscription_id)
    .execute(exec)
    .await
}

/// Params for `CountActiveSeats` (Go `CountActiveSeatsParams`).
pub struct CountActiveSeatsParams {
    pub business_id: i64,
}

/// `CountActiveSeats`: `one`.
pub async fn count_active_seats<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CountActiveSeatsParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(DISTINCT user_id)\nFROM (\n    SELECT b.owner_id AS user_id\n    FROM businesses b\n    JOIN users u ON u.id = b.owner_id\n    WHERE b.id = ? AND COALESCE(b.is_active, TRUE) = TRUE AND COALESCE(u.is_banned, FALSE) = FALSE\n    UNION\n    SELECT member.user_id AS user_id\n    FROM authorization_business_members member\n    JOIN businesses b ON b.id = member.business_id\n    JOIN users u ON u.id = member.user_id\n    WHERE member.business_id = ? AND COALESCE(b.is_active, TRUE) = TRUE\n      AND member.status = 'active' AND COALESCE(u.is_banned, FALSE) = FALSE\n) active_users")
        .bind(params.business_id)
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// `ExpireTrialSubscriptions`: `exec`.
pub async fn expire_trial_subscriptions<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE subscriptions\nSET status = 'expired', updated_at = CURRENT_TIMESTAMP\nWHERE status = 'trial'\n  AND trial_ends_at IS NOT NULL\n  AND trial_ends_at <= CURRENT_TIMESTAMP")
        .execute(exec)
        .await
}

/// Params for `GetActiveManualGrantExpiresAt` (Go `GetActiveManualGrantExpiresAtParams`).
pub struct GetActiveManualGrantExpiresAtParams {
    pub business_id: i64,
    pub plan_id: i64,
}

/// `GetActiveManualGrantExpiresAt`: `one`.
pub async fn get_active_manual_grant_expires_at<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetActiveManualGrantExpiresAtParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT expires_at FROM manual_access_grants\nWHERE business_id = ? AND plan_id = ? AND revoked_at IS NULL\n  AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP)\nORDER BY created_at DESC\nLIMIT 1")
        .bind(params.business_id)
        .bind(params.plan_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetBillingHistoryPeriods` (Go `GetBillingHistoryPeriodsParams`).
pub struct GetBillingHistoryPeriodsParams {
    pub business_id: i64,
    pub billing_month: chrono::NaiveDate,
    pub limit: i32,
}

/// `GetBillingHistoryPeriods`: `many`.
pub async fn get_billing_history_periods<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetBillingHistoryPeriodsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT billing_month, seat_count, price_per_seat_cents, period_start, period_end\nFROM seat_periods\nWHERE business_id = ? AND billing_month < ?\nORDER BY billing_month DESC, period_start ASC\nLIMIT ?")
        .bind(params.business_id)
        .bind(params.billing_month)
        .bind(params.limit)
        .execute(exec)
        .await
}

/// Params for `GetBusinessApps` (Go `GetBusinessAppsParams`).
pub struct GetBusinessAppsParams {
    pub business_id: i64,
}

/// `GetBusinessApps`: `one`.
pub async fn get_business_apps<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetBusinessAppsParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT p.apps\nFROM subscriptions s\nJOIN plans p ON p.id = s.plan_id\nWHERE s.business_id = ?")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetBusinessIDByPaddleSubscription` (Go `GetBusinessIDByPaddleSubscriptionParams`).
pub struct GetBusinessIDByPaddleSubscriptionParams {
    pub paddle_subscription_id: Option<String>,
}

/// `GetBusinessIDByPaddleSubscription`: `one`.
pub async fn get_business_id_by_paddle_subscription<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetBusinessIDByPaddleSubscriptionParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT business_id FROM subscriptions WHERE paddle_subscription_id = ? ORDER BY id ASC LIMIT 1")
        .bind(&params.paddle_subscription_id)
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
    sqlx::query("SELECT id, name, slug, apps, price_per_seat_cents,\n       COALESCE(annual_price_per_seat_cents, 0) AS annual_price_per_seat_cents, trial_days,\n       COALESCE(paddle_price_id, '') AS paddle_price_id, COALESCE(tagline, '') AS tagline,\n       max_seats, max_locations,\n       COALESCE(ai_credits_per_month, 0) AS ai_credits_per_month, COALESCE(features, JSON_ARRAY()) AS features,\n       COALESCE(tier_order, 100) AS tier_order, COALESCE(is_solo, FALSE) AS is_solo, COALESCE(is_featured, FALSE) AS is_featured,\n       is_active, created_at\nFROM plans WHERE id = ?")
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
    sqlx::query("SELECT id, name, slug, apps, price_per_seat_cents,\n       COALESCE(annual_price_per_seat_cents, 0) AS annual_price_per_seat_cents, trial_days,\n       COALESCE(paddle_price_id, '') AS paddle_price_id, COALESCE(tagline, '') AS tagline,\n       max_seats, max_locations,\n       COALESCE(ai_credits_per_month, 0) AS ai_credits_per_month, COALESCE(features, JSON_ARRAY()) AS features,\n       COALESCE(tier_order, 100) AS tier_order, COALESCE(is_solo, FALSE) AS is_solo, COALESCE(is_featured, FALSE) AS is_featured,\n       is_active, created_at\nFROM plans WHERE slug = ? AND is_active = TRUE")
        .bind(&params.slug)
        .fetch_optional(exec)
        .await
}

/// Params for `GetSeatPeriodsForMonth` (Go `GetSeatPeriodsForMonthParams`).
pub struct GetSeatPeriodsForMonthParams {
    pub business_id: i64,
    pub billing_month: chrono::NaiveDate,
}

/// `GetSeatPeriodsForMonth`: `many`.
pub async fn get_seat_periods_for_month<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetSeatPeriodsForMonthParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, subscription_id, business_id, seat_count, price_per_seat_cents,\n       period_start, period_end, billing_month\nFROM seat_periods\nWHERE business_id = ? AND billing_month = ?\nORDER BY period_start ASC")
        .bind(params.business_id)
        .bind(params.billing_month)
        .execute(exec)
        .await
}

/// Params for `GetSubscription` (Go `GetSubscriptionParams`).
pub struct GetSubscriptionParams {
    pub business_id: i64,
}

/// `GetSubscription`: `one`.
pub async fn get_subscription<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetSubscriptionParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT s.id, s.business_id, s.plan_id, p.name AS plan_name, p.slug AS plan_slug,\n       COALESCE(p.ai_credits_per_month, 0) AS ai_credits_per_month, p.apps,\n       p.price_per_seat_cents, s.status, s.seat_count, s.trial_ends_at, s.next_billing_date,\n       s.current_period_ends_at, s.paddle_subscription_id, s.paddle_customer_id, s.created_at\nFROM subscriptions s\nJOIN plans p ON p.id = s.plan_id\nWHERE s.business_id = ?")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `InsertPaddleEventIfNew` (Go `InsertPaddleEventIfNewParams`).
pub struct InsertPaddleEventIfNewParams {
    pub event_id: String,
    pub paddle_subscription_id: String,
    pub occurred_at: chrono::NaiveDate,
}

/// `InsertPaddleEventIfNew`: `execresult`.
pub async fn insert_paddle_event_if_new<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsertPaddleEventIfNewParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT IGNORE INTO paddle_webhook_events (event_id, paddle_subscription_id, occurred_at)\nVALUES (?, ?, ?)")
        .bind(&params.event_id)
        .bind(&params.paddle_subscription_id)
        .bind(params.occurred_at)
        .execute(exec)
        .await
}

/// `ListPlans`: `many`.
pub async fn list_plans<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, name, slug, apps, price_per_seat_cents,\n       COALESCE(annual_price_per_seat_cents, 0) AS annual_price_per_seat_cents, trial_days,\n       COALESCE(paddle_price_id, '') AS paddle_price_id, COALESCE(tagline, '') AS tagline,\n       max_seats, max_locations,\n       COALESCE(ai_credits_per_month, 0) AS ai_credits_per_month, COALESCE(features, JSON_ARRAY()) AS features,\n       COALESCE(tier_order, 100) AS tier_order, COALESCE(is_solo, FALSE) AS is_solo, COALESCE(is_featured, FALSE) AS is_featured,\n       is_active, created_at\nFROM plans\nWHERE is_active = TRUE\nORDER BY is_solo ASC, tier_order ASC, id ASC\nLIMIT 200")
        .execute(exec)
        .await
}

/// Params for `MaxPaddleEventOccurredAt` (Go `MaxPaddleEventOccurredAtParams`).
pub struct MaxPaddleEventOccurredAtParams {
    pub paddle_subscription_id: String,
}

/// `MaxPaddleEventOccurredAt`: `one`.
pub async fn max_paddle_event_occurred_at<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &MaxPaddleEventOccurredAtParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT MAX(occurred_at) AS last_occurred FROM paddle_webhook_events WHERE paddle_subscription_id = ?")
        .bind(&params.paddle_subscription_id)
        .fetch_optional(exec)
        .await
}

/// Params for `OpenSeatPeriod` (Go `OpenSeatPeriodParams`).
pub struct OpenSeatPeriodParams {
    pub subscription_id: i64,
    pub business_id: i64,
    pub seat_count: i32,
    pub price_per_seat_cents: i32,
    pub period_start: chrono::NaiveDate,
    pub billing_month: chrono::NaiveDate,
}

/// `OpenSeatPeriod`: `exec`.
pub async fn open_seat_period<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &OpenSeatPeriodParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO seat_periods (subscription_id, business_id, seat_count, price_per_seat_cents, period_start, billing_month)\nVALUES (?, ?, ?, ?, ?, ?)")
        .bind(params.subscription_id)
        .bind(params.business_id)
        .bind(params.seat_count)
        .bind(params.price_per_seat_cents)
        .bind(params.period_start)
        .bind(params.billing_month)
        .execute(exec)
        .await
}

/// Params for `UpdatePaddleSubscription` (Go `UpdatePaddleSubscriptionParams`).
pub struct UpdatePaddleSubscriptionParams {
    pub paddle_subscription_id: Option<String>,
    pub paddle_customer_id: Option<String>,
    pub status: String,
    pub business_id: i64,
}

/// `UpdatePaddleSubscription`: `execresult`.
pub async fn update_paddle_subscription<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdatePaddleSubscriptionParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE subscriptions\nSET paddle_subscription_id = ?, paddle_customer_id = ?, status = ?, updated_at = CURRENT_TIMESTAMP\nWHERE business_id = ?")
        .bind(&params.paddle_subscription_id)
        .bind(&params.paddle_customer_id)
        .bind(&params.status)
        .bind(params.business_id)
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
}

/// `UpsertSubscription`: `exec`.
pub async fn upsert_subscription<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpsertSubscriptionParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO subscriptions (business_id, plan_id, status, seat_count, trial_ends_at, next_billing_date)\nVALUES (?, ?, ?, ?, ?, ?)\nON DUPLICATE KEY UPDATE\n    plan_id = VALUES(plan_id), status = VALUES(status), seat_count = VALUES(seat_count),\n    trial_ends_at = VALUES(trial_ends_at), updated_at = CURRENT_TIMESTAMP")
        .bind(params.business_id)
        .bind(params.plan_id)
        .bind(&params.status)
        .bind(params.seat_count)
        .bind(params.trial_ends_at)
        .bind(params.next_billing_date)
        .execute(exec)
        .await
}
