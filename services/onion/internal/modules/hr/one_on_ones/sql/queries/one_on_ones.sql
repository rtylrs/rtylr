-- name: ListOneOnOnes :many
SELECT id, owner_id, business_id, manager_id, staff_id, scheduled_at, notes, action_items, status, created_at
FROM hr_one_on_ones
WHERE business_id = ?
ORDER BY scheduled_at DESC
LIMIT 200;

-- name: CreateOneOnOne :execresult
INSERT INTO hr_one_on_ones (owner_id, business_id, manager_id, staff_id, scheduled_at, notes, action_items, status)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);
