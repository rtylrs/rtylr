-- Live repair for installations created before stock-take count presence was
-- explicit. The canonical migration is erp/migrations/008; this file is also
-- traversed by the main schema runner used for existing deployments.
SET @erp_stock_take_needs_count_presence := (
    SELECT COUNT(*) = 2 AND SUM(IS_NULLABLE = 'NO') > 0
    FROM INFORMATION_SCHEMA.COLUMNS
    WHERE TABLE_SCHEMA = DATABASE()
      AND TABLE_NAME = 'erp_stock_take_items'
      AND COLUMN_NAME IN ('counted_qty', 'variance')
);

SET @sql := IF(
    @erp_stock_take_needs_count_presence = 1,
    'UPDATE erp_stock_take_items i JOIN erp_stock_takes t ON t.id = i.stock_take_id SET i.counted_qty = NULL, i.variance = NULL WHERE t.status = ''draft''',
    'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql := IF(
    @erp_stock_take_needs_count_presence = 1,
    'ALTER TABLE erp_stock_take_items MODIFY counted_qty DECIMAL(10,3) NULL, MODIFY variance DECIMAL(10,3) NULL',
    'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
