-- name: ListSetupSteps :many
SELECT id, owner_id, business_id, step_key, completed_at
FROM hr_setup_steps
WHERE business_id = ?
ORDER BY id ASC
LIMIT 500;

-- name: UpsertSetupStep :exec
INSERT INTO hr_setup_steps (owner_id, business_id, step_key, completed_at)
VALUES (?, ?, ?, ?)
ON DUPLICATE KEY UPDATE completed_at = VALUES(completed_at);

-- name: GetHRConfig :one
SELECT id, business_id, setup_completed_at, created_at, updated_at
FROM hr_configs
WHERE business_id = ?
LIMIT 1;

-- name: UpsertHRConfig :exec
INSERT INTO hr_configs (owner_id, business_id, setup_completed_at)
VALUES (sqlc.arg('business_id'), sqlc.arg('business_id'), sqlc.arg('setup_completed_at'))
ON DUPLICATE KEY UPDATE setup_completed_at = VALUES(setup_completed_at), updated_at = NOW();
