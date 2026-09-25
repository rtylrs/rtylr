-- name: ListModifiersByOwner :many
SELECT id, owner_id, business_id, name, price_delta, group_name, is_required, created_at, updated_at
FROM pos_modifiers WHERE business_id = ? ORDER BY group_name, name;

-- name: CreateModifier :execresult
INSERT INTO pos_modifiers (owner_id, business_id, name, price_delta, group_name, is_required)
VALUES (?, ?, ?, ?, ?, ?);

-- name: UpdateModifier :execresult
UPDATE pos_modifiers SET name = ?, price_delta = ?, group_name = ?, is_required = ?
WHERE id = ? AND business_id = ?;

-- name: DeleteModifier :execresult
DELETE FROM pos_modifiers WHERE id = ? AND business_id = ?;
