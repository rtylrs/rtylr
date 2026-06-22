-- Adds indexed self-host license filtering support.
-- Duplicate-index errors are safe under the migration runner's --force mode.

ALTER TABLE self_host_licenses ADD FULLTEXT INDEX idx_self_host_licenses_search (owner_email, customer_name);
