-- name: GetSetupSteps :many
SELECT id, owner_id, business_id, step_key, step_label, is_completed, completed_at, created_at, updated_at
FROM hr_setup_steps
WHERE business_id = ?
ORDER BY id ASC;

-- name: CompleteSetupStep :exec
UPDATE hr_setup_steps
SET is_completed = TRUE, completed_at = NOW(), updated_at = NOW()
WHERE business_id = ? AND step_key = ?;

-- name: GetSetupStepByKey :one
SELECT id, owner_id, business_id, step_key, step_label, is_completed, completed_at, created_at, updated_at
FROM hr_setup_steps
WHERE business_id = ? AND step_key = ?;

-- name: GetHRConfig :one
SELECT h.id, h.owner_id, h.business_id,
       h.fiscal_year_start_month, h.default_currency, h.default_timezone,
       h.default_work_schedule_id, h.overtime_enabled, h.overtime_multiplier,
       h.directory_enabled, h.created_at, h.updated_at
FROM hr_company_config h
WHERE h.business_id = ?;

-- name: UpsertHRConfig :execresult
INSERT INTO hr_company_config (
    owner_id, business_id, fiscal_year_start_month, default_currency,
    default_timezone, default_work_schedule_id, overtime_enabled, overtime_multiplier,
    directory_enabled
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
    fiscal_year_start_month = VALUES(fiscal_year_start_month),
    default_currency = VALUES(default_currency),
    default_timezone = VALUES(default_timezone),
    default_work_schedule_id = VALUES(default_work_schedule_id),
    overtime_enabled = VALUES(overtime_enabled),
    overtime_multiplier = VALUES(overtime_multiplier),
    directory_enabled = VALUES(directory_enabled),
    updated_at = NOW();
