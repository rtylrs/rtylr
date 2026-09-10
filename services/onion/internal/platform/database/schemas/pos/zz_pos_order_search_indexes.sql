-- Adds business-scoped pagination and indexed search support for POS order lists.

SET @migration_table_exists := (
    SELECT COUNT(*)
    FROM information_schema.tables
    WHERE table_schema = DATABASE()
      AND table_name = 'pos_orders'
);

SET @migration_sql := IF(
    @migration_table_exists = 1
    AND NOT EXISTS (
        SELECT 1
        FROM information_schema.statistics
        WHERE table_schema = DATABASE()
          AND table_name = 'pos_orders'
          AND index_name = 'idx_pos_orders_business_status_created'
    ),
    'ALTER TABLE pos_orders ADD INDEX idx_pos_orders_business_status_created (business_id, status, created_at)',
    'SELECT 1'
);
PREPARE migration_statement FROM @migration_sql;
EXECUTE migration_statement;
DEALLOCATE PREPARE migration_statement;

SET @migration_sql := IF(
    @migration_table_exists = 1
    AND NOT EXISTS (
        SELECT 1
        FROM information_schema.statistics
        WHERE table_schema = DATABASE()
          AND table_name = 'pos_orders'
          AND index_name = 'idx_pos_orders_business_session_created'
    ),
    'ALTER TABLE pos_orders ADD INDEX idx_pos_orders_business_session_created (business_id, session_id, created_at)',
    'SELECT 1'
);
PREPARE migration_statement FROM @migration_sql;
EXECUTE migration_statement;
DEALLOCATE PREPARE migration_statement;

SET @migration_sql := IF(
    @migration_table_exists = 1
    AND NOT EXISTS (
        SELECT 1
        FROM information_schema.statistics
        WHERE table_schema = DATABASE()
          AND table_name = 'pos_orders'
          AND index_name = 'idx_pos_orders_business_created'
    ),
    'ALTER TABLE pos_orders ADD INDEX idx_pos_orders_business_created (business_id, created_at)',
    'SELECT 1'
);
PREPARE migration_statement FROM @migration_sql;
EXECUTE migration_statement;
DEALLOCATE PREPARE migration_statement;

SET @migration_sql := IF(
    @migration_table_exists = 1
    AND NOT EXISTS (
        SELECT 1
        FROM information_schema.statistics
        WHERE table_schema = DATABASE()
          AND table_name = 'pos_orders'
          AND index_name = 'idx_pos_orders_search'
    ),
    'ALTER TABLE pos_orders ADD FULLTEXT INDEX idx_pos_orders_search (order_number, customer_name, customer_phone)',
    'SELECT 1'
);
PREPARE migration_statement FROM @migration_sql;
EXECUTE migration_statement;
DEALLOCATE PREPARE migration_statement;
