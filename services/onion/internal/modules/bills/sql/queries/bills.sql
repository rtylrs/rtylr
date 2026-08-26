-- Bounded-result exemption: the legacy Caret tool has no pagination fields.
-- Keep its preview capped at 500 rows until that contract supports paging.
-- name: ListBills :many
SELECT id, owner_id, business_id, supplier_id, COALESCE(supplier_name,'') AS supplier_name, purchase_order_id, amount, currency, due_date, status, description, created_at, updated_at
FROM finance_bills
WHERE business_id = ?
ORDER BY due_date ASC, id ASC
LIMIT 500;

-- name: GetBill :one
SELECT id, owner_id, business_id, supplier_id, COALESCE(supplier_name,'') AS supplier_name, purchase_order_id, amount, currency, due_date, status, description, created_at, updated_at
FROM finance_bills
WHERE id = ? AND business_id = ?;

-- name: DeleteBill :execresult
DELETE FROM finance_bills
WHERE id = ? AND business_id = ? AND status <> 'paid';

-- name: CreateBill :execresult
INSERT INTO finance_bills (owner_id, business_id, supplier_id, supplier_name, purchase_order_id, amount, currency, due_date, status, description)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateBill :execresult
UPDATE finance_bills
SET supplier_id = ?, supplier_name = ?, purchase_order_id = ?, amount = ?, currency = ?, due_date = ?, status = ?, description = ?
WHERE id = ? AND business_id = ? AND status = sqlc.arg(current_status) AND status <> 'paid';

-- name: SumBillsByStatus :one
SELECT COALESCE(SUM(CAST(amount AS REAL)), 0) AS total
FROM finance_bills
WHERE business_id = ? AND status = ?;

-- unbounded-ok: one row per status value, bounded by the enum's cardinality.
-- name: CountBillsByStatus :many
SELECT status, COUNT(*) AS cnt
FROM finance_bills
WHERE business_id = ?
  AND created_at >= DATE_SUB(UTC_TIMESTAMP(), INTERVAL 12 MONTH)
GROUP BY status;

-- name: ListBillsPaginated :many
SELECT id, owner_id, business_id, supplier_id, COALESCE(supplier_name,'') AS supplier_name, purchase_order_id, amount, currency, due_date, status, description, created_at, updated_at
FROM finance_bills
WHERE business_id = ?
ORDER BY due_date ASC, id ASC
LIMIT ? OFFSET ?;

-- name: ListBillsByPurchaseOrder :many
-- Settlement view (#851): which bills already reference one order.
SELECT id, owner_id, business_id, supplier_id, COALESCE(supplier_name,'') AS supplier_name, purchase_order_id, amount, currency, due_date, status, description, created_at, updated_at
FROM finance_bills
WHERE business_id = ? AND purchase_order_id = ?
ORDER BY due_date ASC, id ASC
LIMIT 500;

-- name: CountBillsByPurchaseOrder :one
SELECT COUNT(*) FROM finance_bills WHERE business_id = ? AND purchase_order_id = ?;
