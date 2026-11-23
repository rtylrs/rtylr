-- Review link, prep-time estimate, customer-typed numbers and ringtone on the
-- venue; the estimate on each ticket. Every step is guarded so a stopped run can be
-- retried, and nothing is dropped.

SET @pager_col_exists := (
    SELECT COUNT(*) > 0 FROM information_schema.columns
    WHERE table_schema = DATABASE() AND table_name = 'pager_settings' AND column_name = 'review_url'
);
SET @pager_sql := IF(@pager_col_exists, 'SELECT 1',
    'ALTER TABLE pager_settings ADD COLUMN review_url varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL AFTER label_mode');
PREPARE pager_stmt FROM @pager_sql;
EXECUTE pager_stmt;
DEALLOCATE PREPARE pager_stmt;

SET @pager_col_exists := (
    SELECT COUNT(*) > 0 FROM information_schema.columns
    WHERE table_schema = DATABASE() AND table_name = 'pager_settings' AND column_name = 'prep_minutes'
);
SET @pager_sql := IF(@pager_col_exists, 'SELECT 1',
    "ALTER TABLE pager_settings ADD COLUMN prep_minutes smallint NOT NULL DEFAULT '0' AFTER review_url");
PREPARE pager_stmt FROM @pager_sql;
EXECUTE pager_stmt;
DEALLOCATE PREPARE pager_stmt;

SET @pager_col_exists := (
    SELECT COUNT(*) > 0 FROM information_schema.columns
    WHERE table_schema = DATABASE() AND table_name = 'pager_settings' AND column_name = 'customer_numbers'
);
SET @pager_sql := IF(@pager_col_exists, 'SELECT 1',
    "ALTER TABLE pager_settings ADD COLUMN customer_numbers tinyint(1) NOT NULL DEFAULT '0' AFTER prep_minutes");
PREPARE pager_stmt FROM @pager_sql;
EXECUTE pager_stmt;
DEALLOCATE PREPARE pager_stmt;

SET @pager_col_exists := (
    SELECT COUNT(*) > 0 FROM information_schema.columns
    WHERE table_schema = DATABASE() AND table_name = 'pager_settings' AND column_name = 'ringtone'
);
SET @pager_sql := IF(@pager_col_exists, 'SELECT 1',
    "ALTER TABLE pager_settings ADD COLUMN ringtone varchar(16) CHARACTER SET ascii COLLATE ascii_bin NOT NULL DEFAULT 'bell' AFTER customer_numbers");
PREPARE pager_stmt FROM @pager_sql;
EXECUTE pager_stmt;
DEALLOCATE PREPARE pager_stmt;

SET @pager_col_exists := (
    SELECT COUNT(*) > 0 FROM information_schema.columns
    WHERE table_schema = DATABASE() AND table_name = 'pager_tickets' AND column_name = 'expected_ready_at'
);
SET @pager_sql := IF(@pager_col_exists, 'SELECT 1',
    'ALTER TABLE pager_tickets ADD COLUMN expected_ready_at timestamp(3) NULL DEFAULT NULL AFTER acknowledged_at');
PREPARE pager_stmt FROM @pager_sql;
EXECUTE pager_stmt;
DEALLOCATE PREPARE pager_stmt;
