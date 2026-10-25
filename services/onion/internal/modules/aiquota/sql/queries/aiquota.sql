-- name: GetQuotaMode :one
SELECT mode FROM ai_quota_settings WHERE business_id = ?;

-- name: UpsertQuotaMode :exec
INSERT INTO ai_quota_settings (business_id, mode) VALUES (?, ?)
ON DUPLICATE KEY UPDATE mode = VALUES(mode);

-- name: GetMonthlyTokenUsage :one
SELECT tokens_used FROM ai_usage_monthly
WHERE business_id = ? AND user_id = ? AND billing_month = ?;

-- name: AddMonthlyTokenUsage :exec
INSERT INTO ai_usage_monthly (business_id, user_id, billing_month, tokens_used, requests)
VALUES (?, ?, ?, ?, 1)
ON DUPLICATE KEY UPDATE
  tokens_used = tokens_used + VALUES(tokens_used),
  requests = requests + 1;

-- name: EnsureUsageRow :exec
INSERT IGNORE INTO ai_usage_monthly (business_id, user_id, billing_month, tokens_used, requests)
VALUES (?, ?, ?, 0, 0);

-- name: ReserveMonthlyTokens :execresult
-- Conditional reservation: the increment only lands when the resulting total
-- stays within the caller-supplied allowance, making the check-and-add atomic.
-- When allow_overage is true (paid Pro/Scale plans), the allowance cap is not enforced.
UPDATE ai_usage_monthly
SET tokens_used = tokens_used + sqlc.arg('estimate'),
    requests = requests + 1
WHERE business_id = sqlc.arg('business_id')
  AND user_id = sqlc.arg('user_id')
  AND billing_month = sqlc.arg('billing_month')
  AND (sqlc.arg('allow_overage') = TRUE OR tokens_used + sqlc.arg('estimate') <= sqlc.arg('allowance'));

-- name: AdjustMonthlyTokenUsage :exec
-- Post-stream reconciliation of a reserved estimate against actual usage.
-- No request-count bump: the reservation already counted the request.
UPDATE ai_usage_monthly
SET tokens_used = GREATEST(0, CAST(tokens_used AS SIGNED) + sqlc.arg('delta'))
WHERE business_id = sqlc.arg('business_id')
  AND user_id = sqlc.arg('user_id')
  AND billing_month = sqlc.arg('billing_month');
