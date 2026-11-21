-- name: ListBreaks :many
SELECT id, owner_id, business_id, employee_id, staff_id, staff_name, break_type,
       started_at, ended_at, duration_minutes, created_at, updated_at
FROM hr_breaks
WHERE business_id = ?
  AND (sqlc.narg('employee_id') IS NULL OR employee_id = sqlc.narg('employee_id'))
ORDER BY started_at DESC, id DESC
LIMIT ? OFFSET ?;

-- name: GetBreak :one
SELECT id, owner_id, business_id, employee_id, staff_id, staff_name, break_type,
       started_at, ended_at, duration_minutes, created_at, updated_at
FROM hr_breaks
WHERE business_id = ? AND id = ?;

-- name: GetOpenBreakByEmployee :one
SELECT id, owner_id, business_id, employee_id, staff_id, staff_name, break_type,
       started_at, ended_at, duration_minutes, created_at, updated_at
FROM hr_breaks
WHERE business_id = ? AND employee_id = ? AND ended_at IS NULL
ORDER BY started_at DESC, id DESC
LIMIT 1;

-- name: GetOpenBreakByStaff :one
SELECT id, owner_id, business_id, employee_id, staff_id, staff_name, break_type,
       started_at, ended_at, duration_minutes, created_at, updated_at
FROM hr_breaks
WHERE business_id = ? AND staff_id = ? AND ended_at IS NULL
ORDER BY started_at DESC, id DESC
LIMIT 1;

-- name: CreateBreak :execresult
INSERT INTO hr_breaks (
    owner_id, business_id, employee_id, staff_id, staff_name, break_type, started_at
) VALUES (?, ?, ?, ?, ?, ?, NOW());

-- name: EndBreak :execrows
UPDATE hr_breaks
SET ended_at = NOW(),
    duration_minutes = GREATEST(0, TIMESTAMPDIFF(MINUTE, started_at, NOW())),
    updated_at = NOW()
WHERE business_id = ? AND id = ? AND ended_at IS NULL;
