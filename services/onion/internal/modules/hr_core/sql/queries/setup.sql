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
SELECT id, owner_id, business_id, company_name, fiscal_year_start_month, default_currency,
       default_timezone, default_work_schedule_id, overtime_enabled, overtime_multiplier,
       directory_enabled, created_at, updated_at
FROM hr_company_config
WHERE business_id = ?;

-- name: UpsertHRConfig :execresult
INSERT INTO hr_company_config (
    owner_id, business_id, company_name, fiscal_year_start_month, default_currency,
    default_timezone, default_work_schedule_id, overtime_enabled, overtime_multiplier,
    directory_enabled
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
    company_name = VALUES(company_name),
    fiscal_year_start_month = VALUES(fiscal_year_start_month),
    default_currency = VALUES(default_currency),
    default_timezone = VALUES(default_timezone),
    default_work_schedule_id = VALUES(default_work_schedule_id),
    overtime_enabled = VALUES(overtime_enabled),
    overtime_multiplier = VALUES(overtime_multiplier),
    directory_enabled = VALUES(directory_enabled),
    updated_at = NOW();
