-- name: InsertAICacheEvent :exec
INSERT INTO ai_cache_events (
    event_id, occurred_at, outcome, mode, provider, reason, cache_key_preview,
    business_id, user_id, prompt_chars, response_chars, provider_tokens,
    estimated_savings_tokens, latency_ms
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE event_id = event_id;

-- name: ListRecentAICacheEvents :many
SELECT
    occurred_at, outcome, mode, provider, reason, cache_key_preview,
    business_id, user_id, prompt_chars, response_chars, provider_tokens,
    estimated_savings_tokens, latency_ms
FROM ai_cache_events
WHERE occurred_at >= ?
ORDER BY occurred_at DESC
LIMIT ?;
