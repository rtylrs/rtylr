-- cross-schema read from subscriptions + plans — load subscription and plan details for console administration.
-- name: GetSubscriptionByID :one
SELECT s.id, s.business_id, b.name AS business_name, b.owner_id, COALESCE(u.full_name, '') AS owner_name,
       u.email AS owner_email, s.plan_id, p.name AS plan_name,
       s.status, s.seat_count, s.trial_ends_at, s.next_billing_date, s.current_period_ends_at,
       s.paddle_subscription_id, s.paddle_customer_id, s.created_at, s.updated_at
FROM subscriptions s
JOIN plans p ON p.id = s.plan_id
LEFT JOIN businesses b ON b.id = s.business_id
LEFT JOIN users u ON u.id = b.owner_id
WHERE s.id = ?
LIMIT 1;

-- cross-schema read from subscriptions + plans — load the business's subscription and plan details.
-- name: GetSubscriptionByBusiness :one
SELECT s.id, s.business_id, b.name AS business_name, b.owner_id, COALESCE(u.full_name, '') AS owner_name,
       u.email AS owner_email, s.plan_id, p.name AS plan_name,
       s.status, s.seat_count, s.trial_ends_at, s.next_billing_date, s.current_period_ends_at,
       s.paddle_subscription_id, s.paddle_customer_id, s.created_at, s.updated_at
FROM subscriptions s
JOIN plans p ON p.id = s.plan_id
JOIN businesses b ON b.id = s.business_id
LEFT JOIN users u ON u.id = b.owner_id
WHERE s.business_id = ?
LIMIT 1;

-- cross-schema write to subscriptions — create or update console-managed subscription state.
-- name: UpsertSubscription :exec
INSERT INTO subscriptions (
    business_id, plan_id, status, seat_count, trial_ends_at, next_billing_date,
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

-- name: CountActiveSeatsForBusiness :one
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
