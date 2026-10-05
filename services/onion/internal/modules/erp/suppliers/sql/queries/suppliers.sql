-- name: ListSuppliers :many
SELECT id, owner_id, business_id, name, email, phone, address, city, country,
       tax_id, payment_terms, notes, is_active, created_at, updated_at
FROM erp_suppliers
WHERE business_id = ?
ORDER BY name ASC
LIMIT 200;

-- name: ListSuppliersPaginated :many
SELECT id, owner_id, business_id, name, email, phone, address, city, country,
       tax_id, payment_terms, notes, is_active, created_at, updated_at
FROM erp_suppliers
WHERE business_id = ?
ORDER BY name ASC
LIMIT ? OFFSET ?;

-- name: GetSupplierByID :one
SELECT id, owner_id, business_id, name, email, phone, address, city, country,
       tax_id, payment_terms, notes, is_active, created_at, updated_at
FROM erp_suppliers
WHERE business_id = ? AND id = ?
ORDER BY id ASC
LIMIT 1;

-- name: CountSuppliers :one
SELECT COUNT(*) FROM erp_suppliers WHERE business_id = ?;

-- name: CountActiveSuppliers :one
SELECT COUNT(*) FROM erp_suppliers WHERE business_id = ? AND is_active = 1;

-- name: CreateSupplier :execresult
INSERT INTO erp_suppliers
(owner_id, business_id, name, email, phone, address, city, country, tax_id, payment_terms, notes, is_active)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateSupplier :exec
UPDATE erp_suppliers
SET name = ?, email = ?, phone = ?, address = ?, city = ?, country = ?,
    tax_id = ?, payment_terms = ?, notes = ?, is_active = ?
WHERE business_id = ? AND id = ?;

-- name: DeleteSupplier :exec
DELETE FROM erp_suppliers WHERE business_id = ? AND id = ?;
