-- Unified stock model — Slice 1 (backfill; idempotent, re-runnable).
-- Seeds a default native warehouse per tenant, copies the legacy
-- erp_products.stock_quantity scalar into per-location stock_levels, and writes
-- an 'opening' ledger row documenting the migration origin. No code path
-- changes; the scalar remains authoritative for reads.

-- 1. One default native warehouse per tenant that has products.
INSERT INTO stock_locations (owner_id, business_id, kind, source_table, source_id, name, is_default, is_active)
SELECT DISTINCT owner_id, 0, 'warehouse', 'native', 0, 'Default Warehouse', TRUE, TRUE
FROM erp_products
ON DUPLICATE KEY UPDATE source_id = VALUES(source_id);

-- 2. Seed per-location on_hand from the scalar (re-seed is idempotent).
INSERT INTO stock_levels (owner_id, location_id, product_id, on_hand)
SELECT p.owner_id, l.id, p.id, p.stock_quantity
FROM erp_products p
JOIN stock_locations l
  ON l.owner_id = p.owner_id AND l.source_table = 'native' AND l.source_id = 0 AND l.kind = 'warehouse'
ON DUPLICATE KEY UPDATE on_hand = VALUES(on_hand);

-- 3. Opening ledger row per seeded level (only once).
INSERT INTO stock_ledger
  (owner_id, location_id, product_id, entry_type, quantity_change, quantity_before, quantity_after, reference_type, channel, note)
SELECT sl.owner_id, sl.location_id, sl.product_id, 'opening', sl.on_hand, 0, sl.on_hand, 'migration', 'erp', 'opening balance backfill'
FROM stock_levels sl
WHERE NOT EXISTS (
  SELECT 1 FROM stock_ledger g
  WHERE g.owner_id = sl.owner_id AND g.location_id = sl.location_id
    AND g.product_id = sl.product_id AND g.entry_type = 'opening'
);
