-- name: ListBills :many
SELECT id, owner_id, business_id, supplier_id, amount, currency, due_date, status, description, created_at, updated_at
FROM finance_bills
WHERE business_id = ?
ORDER BY due_date ASC
LIMIT 500;

-- name: GetBill :one
SELECT id, owner_id, business_id, supplier_id, amount, currency, due_date, status, description, created_at, updated_at
FROM finance_bills
WHERE id = ? AND business_id = ?;

-- name: DeleteBill :execresult
DELETE FROM finance_bills
WHERE id = ? AND business_id = ?;

-- name: CreateBill :execresult
INSERT INTO finance_bills (owner_id, business_id, supplier_id, amount, currency, due_date, status, description)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateBill :exec
UPDATE finance_bills
SET supplier_id = ?, amount = ?, currency = ?, due_date = ?, status = ?, description = ?
WHERE id = ? AND business_id = ?;

-- name: SumBillsByStatus :one
SELECT COALESCE(SUM(CAST(amount AS REAL)), 0) AS total
FROM finance_bills
WHERE business_id = ? AND status = ?;

-- name: CountBillsByStatus :many
SELECT status, COUNT(*) AS cnt
FROM finance_bills
WHERE business_id = ?
GROUP BY status;

-- name: ListBillsPaginated :many
SELECT id, owner_id, business_id, supplier_id, amount, currency, due_date, status, description, created_at, updated_at
FROM finance_bills
WHERE business_id = ?
ORDER BY due_date ASC
LIMIT ? OFFSET ?;
