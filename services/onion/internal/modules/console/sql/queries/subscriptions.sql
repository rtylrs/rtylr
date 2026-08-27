-- cross-schema read from subscriptions + plans — load subscription and plan details for console administration.
-- name: GetSubscriptionByID :one
SELECT s.id, s.owner_id, COALESCE(u.full_name, '') AS owner_name, s.plan_id, p.name AS plan_name,
       s.status, s.seat_count, s.trial_ends_at, s.next_billing_date, s.current_period_ends_at,
       s.paddle_subscription_id, s.paddle_customer_id, s.created_at, s.updated_at,
       u.email AS owner_email, b.id AS business_id, b.name AS business_name
FROM subscriptions s
JOIN plans p ON p.id = s.plan_id
LEFT JOIN users u ON u.id = s.owner_id
LEFT JOIN businesses b ON b.owner_id = s.owner_id AND COALESCE(b.is_active, TRUE) = TRUE
WHERE s.id = ?
ORDER BY s.id ASC
LIMIT 1;

-- cross-schema read from subscriptions + plans — load the owner's subscription and plan details.
-- name: GetSubscriptionByOwner :one
SELECT s.id, s.owner_id, COALESCE(u.full_name, '') AS owner_name, s.plan_id, p.name AS plan_name,
       s.status, s.seat_count, s.trial_ends_at, s.next_billing_date, s.current_period_ends_at,
       s.paddle_subscription_id, s.paddle_customer_id, s.created_at, s.updated_at,
       u.email AS owner_email, b.id AS business_id, b.name AS business_name
FROM subscriptions s
JOIN plans p ON p.id = s.plan_id
LEFT JOIN users u ON u.id = s.owner_id
LEFT JOIN businesses b ON b.owner_id = s.owner_id AND COALESCE(b.is_active, TRUE) = TRUE
WHERE s.owner_id = ?
ORDER BY s.owner_id ASC
LIMIT 1;

-- cross-schema write to subscriptions — create or update console-managed subscription state.
-- name: UpsertSubscription :exec
INSERT INTO subscriptions (
    owner_id, plan_id, status, seat_count, trial_ends_at, next_billing_date,
    current_period_ends_at, paddle_subscription_id, paddle_customer_id
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
    plan_id = VALUES(plan_id),
    status = VALUES(status),
    seat_count = VALUES(seat_count),
    trial_ends_at = VALUES(trial_ends_at),
    next_billing_date = VALUES(next_billing_date),
    current_period_ends_at = VALUES(current_period_ends_at),
    paddle_subscription_id = VALUES(paddle_subscription_id),
    paddle_customer_id = VALUES(paddle_customer_id),
    updated_at = NOW();

-- cross-schema write to subscriptions — update console-managed subscription state.
-- name: UpdateSubscription :exec
UPDATE subscriptions
SET plan_id = ?, status = ?, seat_count = ?, trial_ends_at = ?, next_billing_date = ?,
    current_period_ends_at = ?, paddle_subscription_id = ?, paddle_customer_id = ?,
    updated_at = NOW()
WHERE id = ?;

-- cross-schema write to subscriptions — remove console-managed subscription state.
-- name: DeleteSubscription :exec
DELETE FROM subscriptions WHERE id = ?;

-- name: CountActiveSeatsForOwner :one
SELECT COUNT(DISTINCT user_id)
FROM (
    SELECT b.owner_id AS user_id
    FROM businesses b
    JOIN users u ON u.id = b.owner_id
    WHERE b.owner_id = ? AND COALESCE(b.is_active, TRUE) = TRUE AND COALESCE(u.is_banned, FALSE) = FALSE
    UNION
    SELECT p.user_id AS user_id
    FROM user_app_permissions p
    JOIN users u ON u.id = p.user_id
    WHERE p.owner_id = ? AND COALESCE(p.is_active, TRUE) = TRUE AND COALESCE(u.is_banned, FALSE) = FALSE
    UNION
    SELECT u.id AS user_id
    FROM users u
    JOIN businesses b ON b.id = u.business_id
    WHERE b.owner_id = ? AND COALESCE(b.is_active, TRUE) = TRUE AND COALESCE(u.is_banned, FALSE) = FALSE
) active_users;
