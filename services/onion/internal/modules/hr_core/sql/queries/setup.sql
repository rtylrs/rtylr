-- name: GetSetupSteps :many
SELECT id, owner_id, step_key, step_label, is_completed, completed_at, created_at, updated_at
FROM hr_setup_steps
WHERE owner_id = ?
ORDER BY id ASC;

-- name: CompleteSetupStep :exec
UPDATE hr_setup_steps
SET is_completed = TRUE, completed_at = NOW(), updated_at = NOW()
WHERE owner_id = ? AND step_key = ?;

-- name: GetSetupStepByKey :one
SELECT id, owner_id, step_key, step_label, is_completed, completed_at, created_at, updated_at
FROM hr_setup_steps
WHERE owner_id = ? AND step_key = ?;

-- name: SeedSetupStep :exec
INSERT INTO hr_setup_steps (owner_id, step_key, step_label)
VALUES (?, ?, ?)
ON DUPLICATE KEY UPDATE step_label = VALUES(step_label);

-- name: GetHRConfig :one
SELECT id, owner_id, company_name, fiscal_year_start_month, default_currency,
       default_timezone, default_work_schedule_id, overtime_enabled, overtime_multiplier,
       created_at, updated_at
FROM hr_company_config
WHERE owner_id = ?;

-- name: UpsertHRConfig :execresult
INSERT INTO hr_company_config (
    owner_id, company_name, fiscal_year_start_month, default_currency,
    default_timezone, default_work_schedule_id, overtime_enabled, overtime_multiplier
) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
    company_name = VALUES(company_name),
    fiscal_year_start_month = VALUES(fiscal_year_start_month),
    default_currency = VALUES(default_currency),
    default_timezone = VALUES(default_timezone),
    default_work_schedule_id = VALUES(default_work_schedule_id),
    overtime_enabled = VALUES(overtime_enabled),
    overtime_multiplier = VALUES(overtime_multiplier),
    updated_at = NOW();
