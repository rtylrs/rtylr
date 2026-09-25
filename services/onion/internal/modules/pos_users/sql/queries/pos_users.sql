-- name: ListPOSUsersByOwner :many
SELECT id, owner_id, business_id, user_id, location_id, name, role, is_active,
       hourly_rate, max_discount_pct, can_void, can_refund,
       can_open_drawer, can_apply_discount, created_by, created_at, updated_at
FROM pos_users WHERE business_id = ? ORDER BY name ASC LIMIT 100;

-- name: ListPOSUsersByLocation :many
SELECT id, owner_id, business_id, user_id, location_id, name, role, is_active,
       hourly_rate, max_discount_pct, can_void, can_refund,
       can_open_drawer, can_apply_discount, created_by, created_at, updated_at
FROM pos_users WHERE business_id = ? AND location_id = ? AND is_active = TRUE ORDER BY name ASC;

-- name: CreatePOSUser :execresult
INSERT INTO pos_users (owner_id, business_id, user_id, location_id, name, pin_hash, role,
    hourly_rate, max_discount_pct, can_void, can_refund,
    can_open_drawer, can_apply_discount, created_by)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdatePOSUser :exec
-- Role and capability fields are projected from Dash authorization and are
-- intentionally excluded from this POS-owned roster mutation.
UPDATE pos_users SET name = ?, is_active = ?, hourly_rate = ?
WHERE business_id = ? AND id = ?;

-- name: UpdatePOSUserPINHash :exec
UPDATE pos_users SET pin_hash = ? WHERE business_id = ? AND id = ?;

-- name: DeactivatePOSUser :exec
UPDATE pos_users SET is_active = FALSE WHERE business_id = ? AND id = ?;

-- name: FindPOSUserForPINVerify :one
SELECT id, owner_id, business_id, user_id, location_id, name, role, is_active,
       hourly_rate, max_discount_pct, can_void, can_refund,
       can_open_drawer, can_apply_discount, created_by, created_at, updated_at
FROM pos_users WHERE business_id = ? AND location_id = ? AND user_id = ? AND is_active = TRUE LIMIT 1;

-- name: FindPOSUserByUserAndLocation :one
SELECT id, owner_id, business_id, user_id, location_id, name, role, is_active,
       hourly_rate, max_discount_pct, can_void, can_refund,
       can_open_drawer, can_apply_discount, created_by, created_at, updated_at
FROM pos_users WHERE business_id = ? AND user_id = ? AND location_id = ? LIMIT 1;

-- name: GetPINHash :one
SELECT pin_hash
FROM pos_users
WHERE business_id = ? AND id = ?;

-- ── Cross-module read of pos_member_locations (same pos schema file) ─────────
-- Both tables live in pos/schema.sql so sqlc resolves this without extra config.

-- name: GetMemberLocationAccess :one
SELECT id, business_id, user_id, location_id, can_manage_pos_users,
       can_manage_terminals, can_view_reports
FROM pos_member_locations WHERE business_id = ? AND user_id = ? AND location_id = ? LIMIT 1;
