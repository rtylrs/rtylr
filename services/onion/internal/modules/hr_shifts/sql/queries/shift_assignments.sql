-- name: ListRosterAssignments :many
SELECT id, owner_id, employee_id, shift_id, date, status, published, created_at, updated_at
FROM hr_shift_assignments
WHERE owner_id = ? AND date >= ? AND date < ?
ORDER BY date ASC, employee_id ASC;

-- name: CreateShiftAssignment :execresult
INSERT INTO hr_shift_assignments (owner_id, employee_id, shift_id, date, status)
VALUES (?, ?, ?, ?, 'scheduled');

-- name: PublishRosterAssignments :exec
UPDATE hr_shift_assignments
SET published = TRUE, updated_at = NOW()
WHERE owner_id = ? AND date >= ? AND date < ?;

-- name: ListEmployeeAssignmentsForDate :many
SELECT id, owner_id, employee_id, shift_id, date, status, published, created_at, updated_at
FROM hr_shift_assignments
WHERE owner_id = ? AND employee_id = ? AND date = ?;

-- name: ListAssignmentsByShiftAndDate :many
SELECT id, owner_id, employee_id, shift_id, date, status, published, created_at, updated_at
FROM hr_shift_assignments
WHERE owner_id = ? AND shift_id = ? AND date = ?;
