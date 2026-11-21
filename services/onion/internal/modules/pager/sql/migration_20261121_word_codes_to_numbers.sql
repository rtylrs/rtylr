-- Pager match numbers: a scanned phone shows a short number such as "427"
-- instead of a word. Additive and guarded so a stopped run can be retried.
-- match_word, match_code and their unique keys stay, so tickets opened before
-- this change keep their codes until they are labelled or expire. The file
-- name sorts after migration_20261121_word_codes.sql, which this builds on.

SET @pager_match_number_exists := (
    SELECT COUNT(*) > 0 FROM information_schema.columns
    WHERE table_schema = DATABASE() AND table_name = 'pager_tickets' AND column_name = 'match_number'
);
SET @pager_sql := IF(@pager_match_number_exists, 'SELECT 1',
    'ALTER TABLE pager_tickets ADD COLUMN match_number varchar(4) CHARACTER SET ascii COLLATE ascii_bin DEFAULT NULL AFTER match_word');
PREPARE pager_stmt FROM @pager_sql;
EXECUTE pager_stmt;
DEALLOCATE PREPARE pager_stmt;

SET @pager_open_number_code_exists := (
    SELECT COUNT(*) > 0 FROM information_schema.columns
    WHERE table_schema = DATABASE() AND table_name = 'pager_tickets' AND column_name = 'open_match_number'
);
SET @pager_sql := IF(@pager_open_number_code_exists, 'SELECT 1',
    "ALTER TABLE pager_tickets ADD COLUMN open_match_number varchar(4) CHARACTER SET ascii COLLATE ascii_bin GENERATED ALWAYS AS (CASE WHEN status IN ('waiting','ready') AND label_kind IS NULL THEN match_number ELSE NULL END) VIRTUAL AFTER open_match_word");
PREPARE pager_stmt FROM @pager_sql;
EXECUTE pager_stmt;
DEALLOCATE PREPARE pager_stmt;

SET @pager_open_number_code_index_exists := (
    SELECT COUNT(*) > 0 FROM information_schema.statistics
    WHERE table_schema = DATABASE() AND table_name = 'pager_tickets' AND index_name = 'uk_pager_tickets_open_match_number'
);
SET @pager_sql := IF(@pager_open_number_code_index_exists, 'SELECT 1',
    'ALTER TABLE pager_tickets ADD UNIQUE KEY uk_pager_tickets_open_match_number (business_id, open_match_number)');
PREPARE pager_stmt FROM @pager_sql;
EXECUTE pager_stmt;
DEALLOCATE PREPARE pager_stmt;
