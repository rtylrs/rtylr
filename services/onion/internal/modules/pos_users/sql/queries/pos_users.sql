-- name: ListPOSUsersByBusiness :many
SELECT id, business_id, location_id, name, pin, role, is_active,
       hourly_rate, max_discount_pct, can_void, can_refund,
       can_open_drawer, can_apply_discount, created_by, created_at, updated_at
FROM pos_users WHERE business_id = ? ORDER BY name ASC;

-- name: ListPOSUsersByLocation :many
SELECT id, business_id, location_id, name, pin, role, is_active,
       hourly_rate, max_discount_pct, can_void, can_refund,
       can_open_drawer, can_apply_discount, created_by, created_at, updated_at
FROM pos_users WHERE business_id = ? AND location_id = ? AND is_active = TRUE ORDER BY name ASC;

-- name: CreatePOSUser :execresult
INSERT INTO pos_users (business_id, location_id, name, pin, role,
    hourly_rate, max_discount_pct, can_void, can_refund,
    can_open_drawer, can_apply_discount, created_by)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdatePOSUser :exec
UPDATE pos_users SET name = ?, pin = ?, role = ?, is_active = ?,
    hourly_rate = ?, max_discount_pct = ?, can_void = ?, can_refund = ?,
    can_open_drawer = ?, can_apply_discount = ?
WHERE business_id = ? AND id = ?;

-- name: DeactivatePOSUser :exec
UPDATE pos_users SET is_active = FALSE WHERE business_id = ? AND id = ?;

-- name: VerifyPIN :one
SELECT id, business_id, location_id, name, pin, role, is_active,
       hourly_rate, max_discount_pct, can_void, can_refund,
       can_open_drawer, can_apply_discount, created_by, created_at, updated_at
FROM pos_users WHERE business_id = ? AND pin = ? AND is_active = TRUE LIMIT 1;
