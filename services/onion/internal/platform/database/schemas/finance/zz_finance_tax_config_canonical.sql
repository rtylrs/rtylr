-- Canonical tax-config model shared by the tax and invoices modules.
-- Duplicate-index errors are safe under the migration runner's --force mode.

UPDATE finance_tax_config
SET name = CASE WHEN name = '' THEN COALESCE(NULLIF(tax_name, ''), 'Default tax') ELSE name END,
    rate = CASE WHEN rate = 0 AND tax_rate IS NOT NULL THEN tax_rate ELSE rate END,
    is_default = TRUE
WHERE is_default = FALSE OR name = '';

ALTER TABLE finance_tax_config
    DROP INDEX owner_id,
    ADD UNIQUE KEY uq_finance_tax_config_business_name_type (business_id, name, type);
