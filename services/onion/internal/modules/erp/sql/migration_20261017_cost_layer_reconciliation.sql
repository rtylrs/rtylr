-- Bring each location's FIFO cost layers back in line with its on-hand stock.
-- Before stock_core valued every inbound movement, only purchase receipts
-- opened layers, so stock that arrived through transfers, returns, counts or
-- adjustments was missing from the valuation, and a second partial receipt
-- opened a layer for stock it never added. Excess layers are trimmed oldest
-- first; missing quantity is opened at the product's cost price.
UPDATE inventory_cost_layers cl
JOIN (
    SELECT ranked.id,
           LEAST(ranked.quantity_remaining,
                 GREATEST(0, ranked.excess - (ranked.running - ranked.quantity_remaining))) AS take
    FROM (
        SELECT l.id, l.quantity_remaining, x.excess,
               SUM(l.quantity_remaining) OVER (
                   PARTITION BY l.owner_id, l.business_id, l.product_id, l.location_id
                   ORDER BY l.received_at, l.id
               ) AS running
        FROM inventory_cost_layers l
        JOIN (
            SELECT c.owner_id, c.business_id, c.product_id, c.location_id,
                   SUM(c.quantity_remaining) - GREATEST(COALESCE(MAX(sl.on_hand), 0), 0) AS excess
            FROM inventory_cost_layers c
            LEFT JOIN stock_levels sl
              ON sl.owner_id = c.owner_id AND sl.business_id = c.business_id
             AND sl.product_id = c.product_id AND sl.location_id = c.location_id
            WHERE c.quantity_remaining > 0
            GROUP BY c.owner_id, c.business_id, c.product_id, c.location_id
        ) x ON x.owner_id = l.owner_id AND x.business_id = l.business_id
           AND x.product_id = l.product_id AND x.location_id = l.location_id
        WHERE l.quantity_remaining > 0 AND x.excess > 0
    ) ranked
) d ON d.id = cl.id
SET cl.quantity_remaining = cl.quantity_remaining - d.take
WHERE d.take > 0;

INSERT INTO inventory_cost_layers
    (owner_id, business_id, product_id, location_id, quantity_remaining, unit_cost, reference_type)
SELECT sl.owner_id, sl.business_id, sl.product_id, sl.location_id,
       sl.on_hand - COALESCE(layers.remaining, 0),
       GREATEST(COALESCE(p.cost_price, 0), 0),
       'reconciliation'
FROM stock_levels sl
JOIN erp_products p ON p.id = sl.product_id AND p.owner_id = sl.owner_id
LEFT JOIN (
    SELECT owner_id, business_id, product_id, location_id, SUM(quantity_remaining) AS remaining
    FROM inventory_cost_layers
    WHERE quantity_remaining > 0
    GROUP BY owner_id, business_id, product_id, location_id
) layers ON layers.owner_id = sl.owner_id AND layers.business_id = sl.business_id
       AND layers.product_id = sl.product_id AND layers.location_id = sl.location_id
WHERE sl.on_hand > COALESCE(layers.remaining, 0);
