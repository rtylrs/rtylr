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

-- name: EnsureUsageRow :exec
INSERT IGNORE INTO ai_usage_monthly (owner_id, user_id, billing_month, tokens_used, requests)
VALUES (?, ?, ?, 0, 0);

-- name: ReserveMonthlyTokens :execresult
-- Conditional reservation: the increment only lands when the resulting total
-- stays within the caller-supplied allowance, making the check-and-add atomic.
UPDATE ai_usage_monthly
SET tokens_used = tokens_used + ?
WHERE owner_id = ? AND user_id = ? AND billing_month = ?
  AND tokens_used + ? <= ?;

-- name: AdjustMonthlyTokenUsage :exec
-- Post-stream reconciliation of a reserved estimate against actual usage.
-- No request-count bump: the reservation already counted the request.
UPDATE ai_usage_monthly
SET tokens_used = tokens_used + ?
WHERE owner_id = ? AND user_id = ? AND billing_month = ?;
