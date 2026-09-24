-- Make stock-ledger idempotency per movement line, not per document.
--
-- Purchase orders, POS orders, transfers, and stock takes can emit several
-- ledger rows for one reference id. The historical key omitted the movement
-- coordinates, so the second line collided with the first. Product and
-- location are both part of the movement identity; retries of the same line
-- remain idempotent while distinct lines are accepted.
--
-- The ALTER is guarded so this file is safe on both a fresh database (the
-- canonical schema already has the widened key) and an existing database
-- carrying the old four-column key from the historical migration.
SET @stock_ledger_reference_columns := (
    SELECT GROUP_CONCAT(COLUMN_NAME ORDER BY SEQ_IN_INDEX SEPARATOR ',')
    FROM INFORMATION_SCHEMA.STATISTICS
    WHERE TABLE_SCHEMA = DATABASE()
      AND TABLE_NAME = 'stock_ledger'
      AND INDEX_NAME = 'uq_ledger_reference'
);

SET @stock_ledger_reference_sql := IF(
    @stock_ledger_reference_columns = 'owner_id,reference_type,reference_id,entry_type,product_id,location_id',
    'SELECT 1',
    IF(
        @stock_ledger_reference_columns IS NULL,
        'ALTER TABLE stock_ledger ADD UNIQUE KEY uq_ledger_reference (owner_id, reference_type, reference_id, entry_type, product_id, location_id)',
        'ALTER TABLE stock_ledger DROP INDEX uq_ledger_reference, ADD UNIQUE KEY uq_ledger_reference (owner_id, reference_type, reference_id, entry_type, product_id, location_id)'
    )
);

PREPARE stock_ledger_reference_stmt FROM @stock_ledger_reference_sql;
EXECUTE stock_ledger_reference_stmt;
DEALLOCATE PREPARE stock_ledger_reference_stmt;
