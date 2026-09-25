-- onion's support_customers queries select customer_id (ListCustomers and
-- GetCustomerByID, added by onion#897 "Serve the support customer CRM link"),
-- but no schemas commit declared the column, so sqlc failed and onion main did
-- not build:
--
--   queries/support_customers.sql:5:45: column "customer_id" does not exist
--
-- Same failure class as schemas#61: Go merged ahead of its schema. This is the
-- second occurrence in a day, and onion has no CI running go build or sqlc,
-- which is why both landed unnoticed.
--
-- Nullable: it is a link to a CRM customer record and not every support customer
-- has one. The query selects it without COALESCE, so sqlc generates
-- sql.NullInt64, which is consistent with a nullable column.
--
-- Own filename plus an INFORMATION_SCHEMA guard: migrate.sh records applied
-- files by name so editing support/schema.sql alone would be a silent no-op on
-- deployed databases, and zz_ files re-run every pass while a FAILED migration
-- is recorded as applied behind a green check.
SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
   WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'support_customers'
     AND COLUMN_NAME = 'customer_id') = 0,
  'ALTER TABLE support_customers ADD COLUMN customer_id BIGINT NULL AFTER phone',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
