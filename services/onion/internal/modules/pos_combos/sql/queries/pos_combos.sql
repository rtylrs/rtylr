-- name: ListCombosByOwner :many
SELECT id, owner_id, business_id, name, description, price, items, is_active, created_at, updated_at
FROM pos_combos WHERE business_id = ? ORDER BY name
LIMIT 200;

-- name: CreateCombo :execresult
INSERT INTO pos_combos (owner_id, business_id, name, description, price, items, is_active)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: UpdateCombo :execresult
UPDATE pos_combos SET name = ?, description = ?, price = ?, items = ?, is_active = ?
WHERE id = ? AND business_id = ?;

-- name: DeleteCombo :execresult
DELETE FROM pos_combos WHERE id = ? AND business_id = ?;
