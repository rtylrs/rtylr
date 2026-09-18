-- Canonical tax-config model shared by the tax and invoices modules.

UPDATE finance_tax_config
SET name = CASE WHEN name = '' THEN COALESCE(NULLIF(tax_name, ''), 'Default tax') ELSE name END,
    rate = CASE WHEN rate = 0 AND tax_rate IS NOT NULL THEN tax_rate ELSE rate END,
    is_default = TRUE
WHERE is_default = FALSE OR name = '';

-- MySQL has no portable DROP INDEX IF EXISTS / ADD UNIQUE IF NOT EXISTS.
SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'finance_tax_config' AND INDEX_NAME = 'owner_id') > 0,
  'ALTER TABLE finance_tax_config DROP INDEX owner_id',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'finance_tax_config' AND INDEX_NAME = 'uq_finance_tax_config_business_name_type') = 0,
  'ALTER TABLE finance_tax_config ADD UNIQUE KEY uq_finance_tax_config_business_name_type (business_id, name, type)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
