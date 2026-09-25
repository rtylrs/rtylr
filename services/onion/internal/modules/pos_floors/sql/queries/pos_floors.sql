-- name: ListFloorsByBranch :many
SELECT id, business_id, branch_id, name, sort_order, is_active, created_at
FROM pos_floors WHERE business_id = ? AND branch_id = ? AND is_active = TRUE ORDER BY sort_order
LIMIT 200;

-- name: CreateFloor :execresult
INSERT INTO pos_floors (business_id, branch_id, name, sort_order, is_active)
VALUES (?, ?, ?, ?, TRUE);

-- name: UpdateFloor :exec
UPDATE pos_floors SET name = ?, sort_order = ?
WHERE id = ? AND business_id = ?;

-- name: SoftDeleteFloor :exec
UPDATE pos_floors SET is_active = FALSE
WHERE id = ? AND business_id = ?;
