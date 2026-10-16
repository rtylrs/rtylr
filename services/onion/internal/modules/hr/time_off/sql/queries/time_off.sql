-- name: ListTimeOffRequests :many
SELECT t.id, t.owner_id, t.business_id, t.staff_id, t.type, t.start_date, t.end_date, t.days_count,
       t.status, t.reason, t.approved_by, t.approved_at, t.notes, t.created_at, t.updated_at,
       COALESCE(TRIM(CONCAT(s.first_name, ' ', s.last_name)), '') AS staff_name
FROM hr_time_off_requests t
LEFT JOIN hr_staff s ON s.id = t.staff_id AND s.business_id = t.business_id
WHERE t.business_id = ?
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
