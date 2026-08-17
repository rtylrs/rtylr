-- name: ListActiveEmployeeIDs :many
SELECT id FROM hr_employees WHERE business_id = ? AND status = 'active' ORDER BY id LIMIT 500;
