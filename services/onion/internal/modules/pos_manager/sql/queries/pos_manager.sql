-- name: ListManagerPINHashes :many
-- Returns every usable manager credential at a business for bcrypt comparison.
-- Rows with no hash are excluded: a manager whose PIN was never migrated to a
-- hash has no usable credential and must not be reported as a match.
SELECT owner_id, staff_id, pin_hash FROM pos_manager_pins WHERE business_id = ? AND pin_hash IS NOT NULL AND pin_hash <> '' ORDER BY id;

-- name: GetUserPasswordHash :one
-- Reads the account credential the owner must re-enter to set a manager PIN.
-- Only the hash leaves the database; the plaintext is never stored.
SELECT id, password_hash FROM users WHERE id = ?;

-- name: UpdateManagerPINHash :execrows
-- Re-hashes an existing manager credential. Returns the number of rows changed,
-- which is zero when the manager has no credential yet, so the caller can fall
-- back to an insert instead of relying on an upsert: pos_manager_pins carries no
-- unique key on (business_id, staff_id), so ON DUPLICATE KEY UPDATE would never
-- fire and would silently stack duplicate credentials for one manager.
UPDATE pos_manager_pins SET pin_hash = ? WHERE business_id = ? AND staff_id = ?;

-- name: InsertManagerPIN :exec
INSERT INTO pos_manager_pins (owner_id, business_id, staff_id, pin_hash) VALUES (?, ?, ?, ?);
