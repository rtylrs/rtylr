-- Adds indexed CRM customer lookup/search support.
-- Duplicate-index errors are safe under the migration runner's --force mode.

ALTER TABLE crm_customers ADD INDEX idx_crm_customers_business_created (business_id, created_at);
ALTER TABLE crm_customers ADD INDEX idx_crm_customers_business_tier (business_id, loyalty_tier);
ALTER TABLE crm_customers ADD FULLTEXT INDEX idx_crm_customers_search (first_name, last_name, email, phone);
