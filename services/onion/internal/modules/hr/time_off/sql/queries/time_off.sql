-- name: ListTimeOffRequests :many
SELECT id, owner_id, business_id, staff_id, type, start_date, end_date, days_count,
       status, reason, approved_by, approved_at, notes, created_at, updated_at
FROM hr_time_off_requests
WHERE business_id = ?
ORDER BY created_at DESC, id DESC
LIMIT 500;

-- name: CreateTimeOffRequest :execresult
INSERT INTO hr_time_off_requests
(owner_id, business_id, staff_id, type, start_date, end_date, days_count, reason, notes)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: ApproveTimeOffRequest :exec
UPDATE hr_time_off_requests
SET status = ?, approved_by = ?, approved_at = NOW(), notes = ?, updated_at = NOW()
WHERE id = ? AND business_id = ?;
