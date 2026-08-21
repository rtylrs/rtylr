-- Keep invoice references separate from export batch markers.
--
-- Older deployments stored Unix epoch milliseconds in billable_invoice_id
-- when a time entry was exported. Move those legacy markers to the new
-- export_batch_id column while preserving billable_invoice_id as NULL, which
-- is the correct not-yet-invoiced state. The threshold identifies the
-- millisecond timestamps written by the old exporter (and avoids treating
-- ordinary invoice IDs as legacy markers).
SET @sql := IF((SELECT COUNT(*)
                FROM INFORMATION_SCHEMA.COLUMNS
                WHERE TABLE_SCHEMA = DATABASE()
                  AND TABLE_NAME = 'hr_time_entries'
                  AND COLUMN_NAME = 'export_batch_id') = 0,
               'ALTER TABLE hr_time_entries ADD COLUMN export_batch_id BIGINT NULL AFTER billable_invoice_id',
               'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

UPDATE hr_time_entries
SET export_batch_id = billable_invoice_id,
    billable_invoice_id = NULL
WHERE billable_invoice_id >= 1000000000000
  AND export_batch_id IS NULL;
