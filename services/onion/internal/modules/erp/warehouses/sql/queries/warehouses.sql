-- name: ListWarehouses :many
SELECT id, owner_id, business_id, name, address, city, country, manager_name,
       capacity, is_default, is_active, created_at, updated_at
FROM erp_warehouses
WHERE business_id = ?
ORDER BY name ASC;

-- name: CreateWarehouse :execresult
INSERT INTO erp_warehouses (owner_id, business_id, name, address, city, country, manager_name, capacity, is_default, is_active)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateWarehouse :exec
UPDATE erp_warehouses
SET name = ?, address = ?, city = ?, country = ?, manager_name = ?,
    capacity = ?, is_default = ?, is_active = ?
WHERE business_id = ? AND id = ?;

-- name: DeleteWarehouse :exec
DELETE FROM erp_warehouses WHERE business_id = ? AND id = ?;

-- name: CountWarehouses :one
SELECT COUNT(*) FROM erp_warehouses WHERE business_id = ?;

-- name: CountActiveWarehouses :one
SELECT COUNT(*) FROM erp_warehouses WHERE business_id = ? AND is_active = 1;

-- name: GetWarehouseByID :one
SELECT id, owner_id, business_id, name, address, city, country, manager_name,
       capacity, is_default, is_active, created_at, updated_at
FROM erp_warehouses
WHERE business_id = ? AND id = ?
LIMIT 1;

-- name: GetWarehouseForUpdate :one
SELECT id, owner_id, business_id, name, address, city, country, manager_name,
       capacity, is_default, is_active, created_at, updated_at
FROM erp_warehouses
WHERE business_id = ? AND id = ?
LIMIT 1 FOR UPDATE;
