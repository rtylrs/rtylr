-- name: ListSegments :many
SELECT id, owner_id, name, description, rules, customer_count, last_calculated_at, created_at, updated_at
FROM crm_segments WHERE owner_id = ? ORDER BY created_at DESC;

-- name: GetSegmentByID :one
SELECT id, owner_id, name, description, rules, customer_count, last_calculated_at, created_at, updated_at
FROM crm_segments WHERE id = ? AND owner_id = ?;

-- name: CreateSegment :execresult
INSERT INTO crm_segments (owner_id, name, description, rules) VALUES (?, ?, ?, ?);

-- name: UpdateSegment :execresult
UPDATE crm_segments SET name = ?, description = ?, rules = ? WHERE id = ? AND owner_id = ?;

-- name: DeleteSegment :execresult
DELETE FROM crm_segments WHERE id = ? AND owner_id = ?;
