-- name: ListPOSStockMovements :many
SELECT
    l.id,
    l.product_id,
    l.entry_type,
    l.quantity_change,
    l.reference_id,
    l.reference_type,
    l.note,
    l.employee_id,
    l.created_at
FROM stock_ledger l
INNER JOIN stock_locations loc
  ON loc.id = l.location_id AND loc.business_id = l.business_id
WHERE l.business_id = ?
  AND loc.source_table = 'pos_locations'
  AND loc.source_id = ?
  AND l.product_id = ?
  AND l.channel = 'pos'
ORDER BY l.created_at DESC, l.id DESC
LIMIT 100;

-- name: UpdateStockLedgerTransferGroup :exec
UPDATE stock_ledger SET transfer_group = ? WHERE id = ? AND business_id = ?;

-- cross-schema read from erp_products — validate inventory-tracked product
-- name: InventoryProductExists :one
SELECT 1 AS exists_flag
FROM erp_products
WHERE business_id = ? AND id = ? AND is_active = TRUE AND track_inventory = TRUE
LIMIT 1;

-- name: GetBusinessOwnerID :one
SELECT owner_id FROM businesses WHERE id = ? AND is_active = TRUE LIMIT 1;

-- name: LastStockLedgerID :one
SELECT id FROM stock_ledger
WHERE business_id = ? AND location_id = ? AND product_id = ?
ORDER BY id DESC
LIMIT 1;

-- cross-schema read from erp_products + stock_core — branch inventory list
-- name: ListBranchInventory :many
SELECT
    p.id AS id,
    p.id AS product_id,
    p.name,
    p.sku,
    p.unit,
    COALESCE(sl.on_hand, 0) AS quantity,
    COALESCE(sl.reorder_point, p.low_stock_threshold) AS reorder_point,
    COALESCE(sl.reorder_quantity, p.reorder_quantity) AS reorder_quantity,
    p.low_stock_threshold,
    COALESCE(sl.updated_at, p.updated_at) AS updated_at
FROM erp_products p
LEFT JOIN stock_locations loc
  ON loc.business_id = p.business_id
 AND loc.source_table = 'pos_locations'
 AND loc.source_id = ?
LEFT JOIN stock_levels sl
  ON sl.business_id = p.business_id
 AND sl.location_id = loc.id
 AND sl.product_id = p.id
WHERE p.business_id = ?
  AND p.is_active = TRUE
  AND p.track_inventory = TRUE
ORDER BY p.name;

-- name: ListLowStockBranchInventory :many
SELECT
    p.id AS id,
    p.id AS product_id,
    p.name,
    p.sku,
    p.unit,
    COALESCE(sl.on_hand, 0) AS quantity,
    COALESCE(sl.reorder_point, p.low_stock_threshold) AS reorder_point,
    COALESCE(sl.reorder_quantity, p.reorder_quantity) AS reorder_quantity,
    p.low_stock_threshold,
    COALESCE(sl.updated_at, p.updated_at) AS updated_at
FROM erp_products p
LEFT JOIN stock_locations loc
  ON loc.business_id = p.business_id
 AND loc.source_table = 'pos_locations'
 AND loc.source_id = ?
LEFT JOIN stock_levels sl
  ON sl.business_id = p.business_id
 AND sl.location_id = loc.id
 AND sl.product_id = p.id
WHERE p.business_id = ?
  AND p.is_active = TRUE
  AND p.track_inventory = TRUE
  AND COALESCE(sl.reorder_point, p.low_stock_threshold) IS NOT NULL
  AND COALESCE(sl.on_hand, 0) <= COALESCE(sl.reorder_point, p.low_stock_threshold)
ORDER BY COALESCE(sl.on_hand, 0) ASC, p.name;

-- name: GetBranchInventoryItem :many
SELECT
    p.id AS id,
    p.id AS product_id,
    p.name,
    p.sku,
    p.unit,
    COALESCE(sl.on_hand, 0) AS quantity,
    COALESCE(sl.reorder_point, p.low_stock_threshold) AS reorder_point,
    COALESCE(sl.reorder_quantity, p.reorder_quantity) AS reorder_quantity,
    p.low_stock_threshold,
    COALESCE(sl.updated_at, p.updated_at) AS updated_at
FROM erp_products p
LEFT JOIN stock_locations loc
  ON loc.business_id = p.business_id
 AND loc.source_table = 'pos_locations'
 AND loc.source_id = ?
LEFT JOIN stock_levels sl
  ON sl.business_id = p.business_id
 AND sl.location_id = loc.id
 AND sl.product_id = p.id
WHERE p.business_id = ?
  AND p.id = ?
  AND p.is_active = TRUE
  AND p.track_inventory = TRUE
LIMIT 1;
