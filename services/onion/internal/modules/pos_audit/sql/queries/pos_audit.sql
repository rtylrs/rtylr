-- name: ListAuditLogs :many
SELECT id, business_id, branch_id, terminal_id, employee_id, action, entity_type, entity_id, details, ip_address, created_at
FROM pos_audit_log
WHERE business_id = ?
  AND (sqlc.narg('action') IS NULL OR action = sqlc.narg('action'))
  AND (sqlc.narg('entity_type') IS NULL OR entity_type = sqlc.narg('entity_type'))
  AND (sqlc.narg('employee_id') IS NULL OR employee_id = sqlc.narg('employee_id'))
  AND (sqlc.narg('start_date') IS NULL OR created_at >= sqlc.narg('start_date'))
  AND (sqlc.narg('end_date') IS NULL OR created_at <= sqlc.narg('end_date'))
  AND (
    sqlc.narg('cursor_created_at') IS NULL
    OR created_at < sqlc.narg('cursor_created_at')
    OR (created_at = sqlc.narg('cursor_created_at') AND id < sqlc.narg('cursor_id'))
  )
ORDER BY created_at DESC, id DESC
LIMIT ?;

-- name: CreateAuditLog :execresult
INSERT INTO pos_audit_log (business_id, branch_id, terminal_id, employee_id, action, entity_type, entity_id, details, ip_address)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);
