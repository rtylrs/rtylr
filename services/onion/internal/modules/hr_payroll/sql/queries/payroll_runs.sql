-- name: ListPayrollRuns :many
SELECT id, owner_id, period_start, period_end, status, created_by, approved_by, exported_at, total_gross, total_net, created_at, updated_at
FROM hr_payroll_runs
WHERE owner_id = ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: CountPayrollRuns :one
SELECT COUNT(*) FROM hr_payroll_runs WHERE owner_id = ?;

-- name: GetPayrollRun :one
SELECT id, owner_id, period_start, period_end, status, created_by, approved_by, exported_at, total_gross, total_net, created_at, updated_at
FROM hr_payroll_runs
WHERE owner_id = ? AND id = ?;

-- name: CreatePayrollRun :execresult
INSERT INTO hr_payroll_runs (owner_id, period_start, period_end, status, created_by, total_gross, total_net)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: UpdatePayrollRunStatus :exec
UPDATE hr_payroll_runs SET status = ? WHERE owner_id = ? AND id = ?;

-- name: SetPayrollRunApproved :exec
UPDATE hr_payroll_runs SET status = 'approved', approved_by = ? WHERE owner_id = ? AND id = ?;

-- name: SetPayrollRunExported :exec
UPDATE hr_payroll_runs SET status = 'exported', exported_at = CURRENT_TIMESTAMP WHERE owner_id = ? AND id = ?;

-- name: GetPayrollRunByPeriod :one
SELECT id, owner_id, period_start, period_end, status, created_by, approved_by, exported_at, total_gross, total_net, created_at, updated_at
FROM hr_payroll_runs
WHERE owner_id = ? AND period_start = ? AND period_end = ?
LIMIT 1;

-- name: ListDistinctPayrollOwnerIDs :many
SELECT DISTINCT owner_id FROM hr_payroll_config;
