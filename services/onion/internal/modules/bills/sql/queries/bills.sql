-- name: ListBills :many
SELECT id, owner_id, supplier_id, amount, currency, due_date, status, description, created_at, updated_at
FROM finance_bills
WHERE owner_id = ?
ORDER BY due_date ASC;

-- name: CreateBill :execresult
INSERT INTO finance_bills (owner_id, supplier_id, amount, currency, due_date, status, description)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: UpdateBill :exec
UPDATE finance_bills
SET supplier_id = ?, amount = ?, currency = ?, due_date = ?, status = ?, description = ?
WHERE id = ? AND owner_id = ?;
