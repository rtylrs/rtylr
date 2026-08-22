-- Invoice numbers are generated as COUNT(*)+1 outside any lock, so concurrent
-- creations can silently mint duplicates. This migration makes
-- (business_id, invoice_number) unique: every non-first duplicate is first
-- suffixed with its own id (preserving history and keeping the rename
-- collision-free), then the key is added. Both steps are guarded so re-runs
-- are no-ops.

UPDATE finance_invoices
SET invoice_number = CONCAT(invoice_number, '-', id)
WHERE id IN (
    SELECT id FROM (
        SELECT dup.id
        FROM finance_invoices dup
        WHERE EXISTS (
            SELECT 1 FROM finance_invoices first_row
            WHERE first_row.business_id = dup.business_id
              AND first_row.invoice_number = dup.invoice_number
              AND first_row.id < dup.id
        )
    ) AS duplicates
) AND NOT EXISTS (
    SELECT 1 FROM INFORMATION_SCHEMA.STATISTICS
    WHERE TABLE_SCHEMA = DATABASE()
      AND TABLE_NAME = 'finance_invoices'
      AND INDEX_NAME = 'uq_finance_invoices_business_number'
);

SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'finance_invoices' AND INDEX_NAME = 'uq_finance_invoices_business_number') = 0,
  'ALTER TABLE finance_invoices ADD UNIQUE KEY uq_finance_invoices_business_number (business_id, invoice_number)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
