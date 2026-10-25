-- name: ListPlans :many
SELECT id, name, slug, apps, price_per_seat_cents,
       COALESCE(annual_price_per_seat_cents, 0) AS annual_price_per_seat_cents, trial_days,
       COALESCE(paddle_price_id, '') AS paddle_price_id, COALESCE(tagline, '') AS tagline,
       max_seats, max_locations,
       COALESCE(ai_credits_per_month, 0) AS ai_credits_per_month, COALESCE(features, JSON_ARRAY()) AS features,
       COALESCE(tier_order, 100) AS tier_order, COALESCE(is_solo, FALSE) AS is_solo, COALESCE(is_featured, FALSE) AS is_featured,
       is_active, created_at
FROM plans
WHERE is_active = TRUE
ORDER BY is_solo ASC, tier_order ASC, id ASC
LIMIT 200;

-- name: GetPlanBySlug :one
SELECT id, name, slug, apps, price_per_seat_cents,
       COALESCE(annual_price_per_seat_cents, 0) AS annual_price_per_seat_cents, trial_days,
       COALESCE(paddle_price_id, '') AS paddle_price_id, COALESCE(tagline, '') AS tagline,
       max_seats, max_locations,
       COALESCE(ai_credits_per_month, 0) AS ai_credits_per_month, COALESCE(features, JSON_ARRAY()) AS features,
       COALESCE(tier_order, 100) AS tier_order, COALESCE(is_solo, FALSE) AS is_solo, COALESCE(is_featured, FALSE) AS is_featured,
       is_active, created_at
FROM plans WHERE slug = ? AND is_active = TRUE;

-- name: GetPlanByID :one
SELECT id, name, slug, apps, price_per_seat_cents,
       COALESCE(annual_price_per_seat_cents, 0) AS annual_price_per_seat_cents, trial_days,
       COALESCE(paddle_price_id, '') AS paddle_price_id, COALESCE(tagline, '') AS tagline,
       max_seats, max_locations,
       COALESCE(ai_credits_per_month, 0) AS ai_credits_per_month, COALESCE(features, JSON_ARRAY()) AS features,
       COALESCE(tier_order, 100) AS tier_order, COALESCE(is_solo, FALSE) AS is_solo, COALESCE(is_featured, FALSE) AS is_featured,
       is_active, created_at
FROM plans WHERE id = ?;

-- name: GetSubscription :one
SELECT s.id, s.business_id, s.plan_id, p.name AS plan_name, p.slug AS plan_slug,
       COALESCE(p.ai_credits_per_month, 0) AS ai_credits_per_month, p.apps,
       p.price_per_seat_cents, s.status, s.seat_count, s.trial_ends_at, s.next_billing_date,
       s.current_period_ends_at, s.paddle_subscription_id, s.paddle_customer_id, s.created_at
FROM subscriptions s
JOIN plans p ON p.id = s.plan_id
WHERE s.business_id = ?;

-- name: GetActiveManualGrantExpiresAt :one
SELECT expires_at FROM manual_access_grants
WHERE business_id = ? AND plan_id = ? AND revoked_at IS NULL
  AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP)
ORDER BY created_at DESC
LIMIT 1;

-- name: UpsertSubscription :exec
INSERT INTO subscriptions (business_id, plan_id, status, seat_count, trial_ends_at, next_billing_date)
VALUES (?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
    plan_id = VALUES(plan_id), status = VALUES(status), seat_count = VALUES(seat_count),
    trial_ends_at = VALUES(trial_ends_at), updated_at = CURRENT_TIMESTAMP;

-- name: CountActiveSeats :one
SELECT COUNT(DISTINCT user_id)
FROM (
    SELECT b.owner_id AS user_id
    FROM businesses b
    JOIN users u ON u.id = b.owner_id
    WHERE b.id = sqlc.arg(business_id) AND COALESCE(b.is_active, TRUE) = TRUE AND COALESCE(u.is_banned, FALSE) = FALSE
    UNION
    SELECT member.user_id AS user_id
    FROM authorization_business_members member
    JOIN businesses b ON b.id = member.business_id
    JOIN users u ON u.id = member.user_id
    WHERE member.business_id = sqlc.arg(business_id) AND COALESCE(b.is_active, TRUE) = TRUE
      AND member.status = 'active' AND COALESCE(u.is_banned, FALSE) = FALSE
) active_users;

-- name: GetBusinessApps :one
SELECT p.apps
FROM subscriptions s
JOIN plans p ON p.id = s.plan_id
WHERE s.business_id = ?;

-- name: ExpireTrialSubscriptions :exec
UPDATE subscriptions
SET status = 'expired', updated_at = CURRENT_TIMESTAMP
WHERE status = 'trial'
  AND trial_ends_at IS NOT NULL
  AND trial_ends_at <= CURRENT_TIMESTAMP;

-- name: CloseActiveSeatPeriod :exec
UPDATE seat_periods SET period_end = ?
WHERE subscription_id = ? AND period_end IS NULL;

-- name: OpenSeatPeriod :exec
INSERT INTO seat_periods (subscription_id, business_id, seat_count, price_per_seat_cents, period_start, billing_month)
VALUES (?, ?, ?, ?, ?, ?);

-- name: GetSeatPeriodsForMonth :many
SELECT id, subscription_id, business_id, seat_count, price_per_seat_cents,
       period_start, period_end, billing_month
FROM seat_periods
WHERE business_id = ? AND billing_month = ?
ORDER BY period_start ASC;

-- name: GetBillingHistoryPeriods :many
SELECT billing_month, seat_count, price_per_seat_cents, period_start, period_end
FROM seat_periods
WHERE business_id = ? AND billing_month < ?
ORDER BY billing_month DESC, period_start ASC
LIMIT ?;

-- name: UpdatePaddleSubscription :execresult
UPDATE subscriptions
SET paddle_subscription_id = ?, paddle_customer_id = ?, status = ?, updated_at = CURRENT_TIMESTAMP
WHERE business_id = ?;

-- name: GetBusinessIDByPaddleSubscription :one
SELECT business_id FROM subscriptions WHERE paddle_subscription_id = ? ORDER BY id ASC LIMIT 1;

-- name: InsertPaddleEventIfNew :execresult
INSERT IGNORE INTO paddle_webhook_events (event_id, paddle_subscription_id, occurred_at)
VALUES (?, ?, ?);

-- name: MaxPaddleEventOccurredAt :one
SELECT MAX(occurred_at) AS last_occurred FROM paddle_webhook_events WHERE paddle_subscription_id = ?;
