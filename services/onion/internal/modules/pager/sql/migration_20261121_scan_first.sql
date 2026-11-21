-- Scan-first pager: a phone's ticket starts unlabelled with a match code, and
-- the cashier labels it with an order number or a name. Every step is guarded
-- so a stopped run can be retried. Nothing is dropped: pager_settings.strict
-- and pager_tickets.source / phone_linked stay but are no longer read.

SET @pager_label_mode_exists := (
    SELECT COUNT(*) > 0 FROM information_schema.columns
    WHERE table_schema = DATABASE() AND table_name = 'pager_settings' AND column_name = 'label_mode'
);
SET @pager_sql := IF(@pager_label_mode_exists, 'SELECT 1',
    "ALTER TABLE pager_settings ADD COLUMN label_mode varchar(8) CHARACTER SET ascii COLLATE ascii_bin NOT NULL DEFAULT 'either' AFTER strict");
PREPARE pager_stmt FROM @pager_sql;
EXECUTE pager_stmt;
DEALLOCATE PREPARE pager_stmt;

-- A name or an unlabelled ticket has no number. NULL leaves open_number NULL,
-- which its unique key allows any number of times.
SET @pager_number_required := (
    SELECT COUNT(*) > 0 FROM information_schema.columns
    WHERE table_schema = DATABASE() AND table_name = 'pager_tickets' AND column_name = 'number'
      AND is_nullable = 'NO'
);
SET @pager_sql := IF(@pager_number_required,
    'ALTER TABLE pager_tickets MODIFY COLUMN number varchar(12) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL',
    'SELECT 1');
PREPARE pager_stmt FROM @pager_sql;
EXECUTE pager_stmt;
DEALLOCATE PREPARE pager_stmt;

SET @pager_label_exists := (
    SELECT COUNT(*) > 0 FROM information_schema.columns
    WHERE table_schema = DATABASE() AND table_name = 'pager_tickets' AND column_name = 'label'
);
SET @pager_sql := IF(@pager_label_exists, 'SELECT 1',
    'ALTER TABLE pager_tickets ADD COLUMN label varchar(40) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL AFTER number');
PREPARE pager_stmt FROM @pager_sql;
EXECUTE pager_stmt;
DEALLOCATE PREPARE pager_stmt;

SET @pager_label_kind_exists := (
    SELECT COUNT(*) > 0 FROM information_schema.columns
    WHERE table_schema = DATABASE() AND table_name = 'pager_tickets' AND column_name = 'label_kind'
);
SET @pager_sql := IF(@pager_label_kind_exists, 'SELECT 1',
    'ALTER TABLE pager_tickets ADD COLUMN label_kind varchar(8) CHARACTER SET ascii COLLATE ascii_bin DEFAULT NULL AFTER label');
PREPARE pager_stmt FROM @pager_sql;
EXECUTE pager_stmt;
DEALLOCATE PREPARE pager_stmt;

SET @pager_match_code_exists := (
    SELECT COUNT(*) > 0 FROM information_schema.columns
    WHERE table_schema = DATABASE() AND table_name = 'pager_tickets' AND column_name = 'match_code'
);
SET @pager_sql := IF(@pager_match_code_exists, 'SELECT 1',
    'ALTER TABLE pager_tickets ADD COLUMN match_code char(2) CHARACTER SET ascii COLLATE ascii_bin DEFAULT NULL AFTER label_kind');
PREPARE pager_stmt FROM @pager_sql;
EXECUTE pager_stmt;
DEALLOCATE PREPARE pager_stmt;

-- Every ticket from before this change was entered by its number.
UPDATE pager_tickets
SET label = number, label_kind = 'number'
WHERE label_kind IS NULL AND number IS NOT NULL;

SET @pager_open_code_exists := (
    SELECT COUNT(*) > 0 FROM information_schema.columns
    WHERE table_schema = DATABASE() AND table_name = 'pager_tickets' AND column_name = 'open_match_code'
);
SET @pager_sql := IF(@pager_open_code_exists, 'SELECT 1',
    "ALTER TABLE pager_tickets ADD COLUMN open_match_code char(2) CHARACTER SET ascii COLLATE ascii_bin GENERATED ALWAYS AS (CASE WHEN status IN ('waiting','ready') AND label_kind IS NULL THEN match_code ELSE NULL END) VIRTUAL AFTER open_number");
PREPARE pager_stmt FROM @pager_sql;
EXECUTE pager_stmt;
DEALLOCATE PREPARE pager_stmt;

SET @pager_open_code_index_exists := (
    SELECT COUNT(*) > 0 FROM information_schema.statistics
    WHERE table_schema = DATABASE() AND table_name = 'pager_tickets' AND index_name = 'uk_pager_tickets_open_match_code'
);
SET @pager_sql := IF(@pager_open_code_index_exists, 'SELECT 1',
    'ALTER TABLE pager_tickets ADD UNIQUE KEY uk_pager_tickets_open_match_code (business_id, open_match_code)');
PREPARE pager_stmt FROM @pager_sql;
EXECUTE pager_stmt;
DEALLOCATE PREPARE pager_stmt;
