-- name: ListPayrollRuns :many
SELECT id, owner_id, business_id, period_start, period_end, status, created_by, approved_by, exported_at, total_gross, total_net, created_at, updated_at
FROM hr_payroll_runs
WHERE business_id = ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: CountPayrollRuns :one
SELECT COUNT(*) FROM hr_payroll_runs WHERE business_id = ?;

-- name: GetPayrollRun :one
SELECT id, owner_id, business_id, period_start, period_end, status, created_by, approved_by, exported_at, total_gross, total_net, created_at, updated_at
FROM hr_payroll_runs
WHERE business_id = ? AND id = ?;

-- name: CreatePayrollRun :execresult
INSERT INTO hr_payroll_runs (owner_id, business_id, period_start, period_end, status, created_by, total_gross, total_net)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- Payroll approval currently has no persisted approval-level cursor. Only the
-- draft -> pending_approval transition is valid here; other state changes use
-- dedicated guarded queries below.
-- name: UpdatePayrollRunStatus :execresult
UPDATE hr_payroll_runs
SET status = ?
WHERE business_id = ?
  AND id = ?
  AND status = 'draft'
  AND ? = 'pending_approval';

-- name: SetPayrollRunApproved :execresult
UPDATE hr_payroll_runs
SET status = 'approved', approved_by = ?
WHERE business_id = ? AND id = ? AND status = 'pending_approval';

-- name: SetPayrollRunExported :execresult
UPDATE hr_payroll_runs
SET status = 'exported', exported_at = CURRENT_TIMESTAMP
WHERE business_id = ? AND id = ? AND status = 'exporting' AND exported_at IS NULL;

-- name: ClaimPayrollRunForExport :execresult
UPDATE hr_payroll_runs
SET status = 'exporting'
WHERE business_id = ? AND id = ? AND status = 'approved' AND exported_at IS NULL;

-- name: ReleasePayrollRunExportClaim :exec
UPDATE hr_payroll_runs
SET status = 'approved'
WHERE business_id = ? AND id = ? AND status = 'exporting' AND exported_at IS NULL;

-- name: GetPayrollRunByPeriod :one
SELECT id, owner_id, business_id, period_start, period_end, status, created_by, approved_by, exported_at, total_gross, total_net, created_at, updated_at
FROM hr_payroll_runs
WHERE business_id = ? AND period_start = ? AND period_end = ?
ORDER BY id ASC
LIMIT 1;

-- ListDistinctPayrollOwnerIDs scans every business with a payroll config, not
-- just one business, so it stays unpaginated by design - bounded by the same
-- generous hard LIMIT as the other cross-tenant enumeration queries in this
-- module (see ListAllPayrollConfigs).
-- name: ListDistinctPayrollOwnerIDs :many
SELECT DISTINCT business_id FROM hr_payroll_config
ORDER BY business_id
LIMIT 5000;

-- name: ListPayrollRunsForEmployee :many
SELECT r.id, r.owner_id, r.business_id, r.period_start, r.period_end, r.status, r.created_by, r.approved_by, r.exported_at, r.total_gross, r.total_net, r.created_at, r.updated_at
FROM hr_payroll_runs r
JOIN hr_payroll_entries e ON e.payroll_run_id = r.id AND e.business_id = r.business_id
WHERE r.business_id = ? AND e.employee_id = ? AND r.status IN (?, ?)
ORDER BY r.created_at DESC
LIMIT ? OFFSET ?;

-- name: CountPayrollRunsForEmployee :one
SELECT COUNT(*) FROM hr_payroll_runs r
JOIN hr_payroll_entries e ON e.payroll_run_id = r.id AND e.business_id = r.business_id
WHERE r.business_id = ? AND e.employee_id = ? AND r.status IN (?, ?);
