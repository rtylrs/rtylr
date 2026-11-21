-- Pager match words: a scanned phone shows a word such as "panda" instead of
-- a two character code. Additive and guarded so a stopped run can be retried.
-- match_code and its unique key stay, so tickets opened before this change
-- keep their codes until they are labelled or expire.

SET @pager_match_word_exists := (
    SELECT COUNT(*) > 0 FROM information_schema.columns
    WHERE table_schema = DATABASE() AND table_name = 'pager_tickets' AND column_name = 'match_word'
);
SET @pager_sql := IF(@pager_match_word_exists, 'SELECT 1',
    'ALTER TABLE pager_tickets ADD COLUMN match_word varchar(16) CHARACTER SET ascii COLLATE ascii_bin DEFAULT NULL AFTER match_code');
PREPARE pager_stmt FROM @pager_sql;
EXECUTE pager_stmt;
DEALLOCATE PREPARE pager_stmt;

SET @pager_open_word_exists := (
    SELECT COUNT(*) > 0 FROM information_schema.columns
    WHERE table_schema = DATABASE() AND table_name = 'pager_tickets' AND column_name = 'open_match_word'
);
SET @pager_sql := IF(@pager_open_word_exists, 'SELECT 1',
    "ALTER TABLE pager_tickets ADD COLUMN open_match_word varchar(16) CHARACTER SET ascii COLLATE ascii_bin GENERATED ALWAYS AS (CASE WHEN status IN ('waiting','ready') AND label_kind IS NULL THEN match_word ELSE NULL END) VIRTUAL AFTER open_match_code");
PREPARE pager_stmt FROM @pager_sql;
EXECUTE pager_stmt;
DEALLOCATE PREPARE pager_stmt;

SET @pager_open_word_index_exists := (
    SELECT COUNT(*) > 0 FROM information_schema.statistics
    WHERE table_schema = DATABASE() AND table_name = 'pager_tickets' AND index_name = 'uk_pager_tickets_open_match_word'
);
SET @pager_sql := IF(@pager_open_word_index_exists, 'SELECT 1',
    'ALTER TABLE pager_tickets ADD UNIQUE KEY uk_pager_tickets_open_match_word (business_id, open_match_word)');
PREPARE pager_stmt FROM @pager_sql;
EXECUTE pager_stmt;
DEALLOCATE PREPARE pager_stmt;
