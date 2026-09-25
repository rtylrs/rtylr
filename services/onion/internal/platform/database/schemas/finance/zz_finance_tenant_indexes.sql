-- Business-leading finance indexes for bill queues, report totals, bank
-- balance aggregation, and journal history/status reporting. Roll back
-- individual keys only after checking live plans with DROP INDEX <name> ON the
-- table; existing owner-leading indexes remain unchanged.
SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'finance_bills' AND INDEX_NAME = 'idx_finance_bills_business_due') = 0,
  'ALTER TABLE finance_bills ADD INDEX idx_finance_bills_business_due (business_id, due_date, id)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'finance_bills' AND INDEX_NAME = 'idx_finance_bills_business_status') = 0,
  'ALTER TABLE finance_bills ADD INDEX idx_finance_bills_business_status (business_id, status)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'finance_bank_accounts' AND INDEX_NAME = 'idx_finance_bank_accounts_business') = 0,
  'ALTER TABLE finance_bank_accounts ADD INDEX idx_finance_bank_accounts_business (business_id)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'finance_journal_entries' AND INDEX_NAME = 'idx_finance_journal_entries_business_date') = 0,
  'ALTER TABLE finance_journal_entries ADD INDEX idx_finance_journal_entries_business_date (business_id, date, id)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'finance_journal_entries' AND INDEX_NAME = 'idx_finance_journal_entries_business_status') = 0,
  'ALTER TABLE finance_journal_entries ADD INDEX idx_finance_journal_entries_business_status (business_id, status)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
