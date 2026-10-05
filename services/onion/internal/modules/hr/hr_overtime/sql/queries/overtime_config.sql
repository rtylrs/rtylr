-- name: GetOvertimeConfig :one
SELECT id, owner_id, business_id, daily_threshold_minutes, default_multiplier, auto_detect_enabled,
       created_at, updated_at
FROM hr_overtime_config
WHERE business_id = ?;

-- name: UpsertOvertimeConfig :execresult
INSERT INTO hr_overtime_config (owner_id, business_id, daily_threshold_minutes, default_multiplier, auto_detect_enabled)
VALUES (?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE daily_threshold_minutes = VALUES(daily_threshold_minutes),
    default_multiplier = VALUES(default_multiplier), auto_detect_enabled = VALUES(auto_detect_enabled),
    updated_at = NOW();

-- name: ListAllConfiguredOwnerIDs :many
SELECT DISTINCT business_id FROM hr_overtime_config WHERE auto_detect_enabled = TRUE
ORDER BY business_id ASC
LIMIT ? OFFSET ?;
