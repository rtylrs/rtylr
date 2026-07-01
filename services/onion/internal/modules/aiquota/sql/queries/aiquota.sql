-- cross-schema read from subscriptions + plans (owned by subscriptions module)
-- name: GetSubscriptionEntitlement :one
SELECT p.slug, p.ai_credits_per_month, s.seat_count, s.status
FROM subscriptions s
JOIN plans p ON p.id = s.plan_id
WHERE s.owner_id = ?;

-- cross-schema read from plans — free tier fallback
-- name: GetFreePlanCredits :one
SELECT ai_credits_per_month FROM plans WHERE slug = 'free' AND is_active = 1 LIMIT 1;

-- name: GetQuotaMode :one
SELECT mode FROM ai_quota_settings WHERE owner_id = ?;

-- name: UpsertQuotaMode :exec
INSERT INTO ai_quota_settings (owner_id, mode) VALUES (?, ?)
ON DUPLICATE KEY UPDATE mode = VALUES(mode);

-- name: GetMonthlyTokenUsage :one
SELECT tokens_used FROM ai_usage_monthly
WHERE owner_id = ? AND user_id = ? AND billing_month = ?;

-- name: AddMonthlyTokenUsage :exec
INSERT INTO ai_usage_monthly (owner_id, user_id, billing_month, tokens_used, requests)
VALUES (?, ?, ?, ?, 1)
ON DUPLICATE KEY UPDATE
  tokens_used = tokens_used + VALUES(tokens_used),
  requests = requests + 1;
