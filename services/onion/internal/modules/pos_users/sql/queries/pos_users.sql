-- name: ListPOSUsersByOwner :many
SELECT id, owner_id, location_id, name, pin, pin_hash, role, is_active,
       hourly_rate, max_discount_pct, can_void, can_refund,
       can_open_drawer, can_apply_discount, created_by, created_at, updated_at
FROM pos_users WHERE owner_id = ? ORDER BY name ASC;

-- name: ListPOSUsersByLocation :many
SELECT id, owner_id, location_id, name, pin, pin_hash, role, is_active,
       hourly_rate, max_discount_pct, can_void, can_refund,
       can_open_drawer, can_apply_discount, created_by, created_at, updated_at
FROM pos_users WHERE owner_id = ? AND location_id = ? AND is_active = TRUE ORDER BY name ASC;

-- name: CreatePOSUser :execresult
INSERT INTO pos_users (owner_id, location_id, name, pin, pin_hash, role,
    hourly_rate, max_discount_pct, can_void, can_refund,
    can_open_drawer, can_apply_discount, created_by)
VALUES (?, ?, ?, '', ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdatePOSUser :exec
UPDATE pos_users SET name = ?, role = ?, is_active = ?,
    hourly_rate = ?, max_discount_pct = ?, can_void = ?, can_refund = ?,
    can_open_drawer = ?, can_apply_discount = ?
WHERE owner_id = ? AND id = ?;

-- name: UpdatePOSUserPINHash :exec
UPDATE pos_users SET pin_hash = ?, pin = '' WHERE owner_id = ? AND id = ?;

-- name: DeactivatePOSUser :exec
UPDATE pos_users SET is_active = FALSE WHERE owner_id = ? AND id = ?;

-- name: FindPOSUserForPINVerify :one
SELECT id, owner_id, location_id, name, pin, pin_hash, role, is_active,
       hourly_rate, max_discount_pct, can_void, can_refund,
       can_open_drawer, can_apply_discount, created_by, created_at, updated_at
FROM pos_users WHERE owner_id = ? AND name = ? AND is_active = TRUE LIMIT 1;
