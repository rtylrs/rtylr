-- name: VerifyManagerPIN :one
SELECT staff_id FROM pos_manager_pins WHERE owner_id = ? AND pin = ?;
