-- Restores two columns that onion's queries require and no reachable schemas
-- commit declared. onion pinned a schemas commit that lived only on a branch
-- which has since been deleted, so the commit is unreachable and the columns
-- went with it. sqlc therefore failed and onion did not build:
--
--   bills/queries/bills.sql:4:93:    column "purchase_order_id" does not exist
--   invoices/queries/invoices.sql:2:109: column "customer_id" does not exist
--
-- Found by fixing one and re-running: sqlc stops at the first failing module
-- (bills was 82 of 132), so the invoices gap was invisible until bills passed.
-- After both, all 132 modules generate and `go build ./...` exits 0, which is
-- what establishes that these two are the whole gap rather than the first two.
--
-- Both BIGINT NULL, settled by onion rather than guessed:
--   * finance_bills.purchase_order_id maps through nullableInt64 /
--     nullableInt64FromInput, so sqlc must produce sql.NullInt64. A bill need
--     not originate from a purchase order.
--   * finance_invoices.customer_id maps through nullableInt64FromPtr /
--     nullableInt64Ptr, likewise. An invoice need not be linked to a customer.
--
-- The CREATE TABLEs in finance/schema.sql are updated too, because sqlc reads
-- those rather than the migrations. This file is what brings already-deployed
-- databases forward.
SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
   WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'finance_bills'
     AND COLUMN_NAME = 'purchase_order_id') = 0,
  'ALTER TABLE finance_bills ADD COLUMN purchase_order_id BIGINT NULL',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql2 := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
   WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'finance_invoices'
     AND COLUMN_NAME = 'customer_id') = 0,
  'ALTER TABLE finance_invoices ADD COLUMN customer_id BIGINT NULL',
  'SELECT 1'
);
PREPARE stmt2 FROM @sql2; EXECUTE stmt2; DEALLOCATE PREPARE stmt2;
