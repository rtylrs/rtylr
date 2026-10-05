-- Preserve the cost basis of stock that existed before receipt layers were added.
-- Future receipts append their actual unit costs; this one-time layer is only
-- the opening balance and is never changed when the product master changes.
INSERT INTO inventory_cost_layers (
    owner_id, business_id, product_id, location_id, quantity_remaining, unit_cost,
    received_at, reference_type
)
SELECT
    sl.owner_id, sl.business_id, sl.product_id, sl.location_id, sl.on_hand,
    p.cost_price, CURRENT_TIMESTAMP, 'opening_balance'
FROM stock_levels sl
JOIN erp_products p
  ON p.id = sl.product_id AND p.business_id = sl.business_id
WHERE sl.on_hand > 0
  AND NOT EXISTS (
      SELECT 1
      FROM inventory_cost_layers cl
      WHERE cl.owner_id = sl.owner_id
        AND cl.product_id = sl.product_id
        AND cl.location_id = sl.location_id
        AND cl.reference_type = 'opening_balance'
  );
