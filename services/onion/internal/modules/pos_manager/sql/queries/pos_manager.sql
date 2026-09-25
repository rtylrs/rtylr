-- name: VerifyManagerPIN :one
SELECT owner_id, staff_id FROM pos_manager_pins WHERE business_id = ? AND pin = ?;
