-- Draft stock-take rows do not yet have physical counts. Preserve in-progress
-- and completed counts, but make legacy draft rows explicitly uncounted before
-- the columns become nullable.
UPDATE erp_stock_take_items i
JOIN erp_stock_takes t ON t.id = i.stock_take_id
SET i.counted_qty = NULL,
    i.variance = NULL
WHERE t.status = 'draft';

ALTER TABLE erp_stock_take_items
    MODIFY counted_qty DECIMAL(10,3) NULL,
    MODIFY variance DECIMAL(10,3) NULL;
