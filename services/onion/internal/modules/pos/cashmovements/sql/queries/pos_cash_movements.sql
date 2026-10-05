-- name: ListCashMovementsBySession :many
SELECT id, business_id, branch_id, terminal_id, session_id, employee_id,
       type, amount, reason, approved_by, created_at
FROM pos_cash_movements
WHERE business_id = ? AND session_id = ?
ORDER BY created_at DESC
LIMIT 200;

-- name: CreateCashMovement :execresult
INSERT INTO pos_cash_movements (business_id, branch_id, terminal_id, session_id, employee_id, type, amount, reason, approved_by)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);
