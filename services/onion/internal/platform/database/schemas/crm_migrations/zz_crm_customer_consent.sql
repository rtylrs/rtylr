-- Preserve explicit CRM communication consent for existing installations.
ALTER TABLE crm_customers ADD COLUMN IF NOT EXISTS consent JSON NULL AFTER notes;
