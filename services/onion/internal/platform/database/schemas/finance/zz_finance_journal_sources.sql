-- Add source identity to journal entries created by external finance producers.
-- NULL values keep legacy/manual entries outside the idempotency key.
SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'finance_journal_entries' AND COLUMN_NAME = 'source_type') = 0,
  'ALTER TABLE finance_journal_entries ADD COLUMN source_type VARCHAR(32) NULL, ADD COLUMN source_id BIGINT NULL',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'finance_journal_entries' AND INDEX_NAME = 'uq_finance_journal_entries_source') = 0,
  'ALTER TABLE finance_journal_entries ADD UNIQUE KEY uq_finance_journal_entries_source (business_id, source_type, source_id)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
