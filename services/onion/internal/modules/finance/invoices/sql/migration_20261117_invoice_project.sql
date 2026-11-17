-- An invoice can be tagged with the Work project it bills, so a project can
-- report what it has invoiced, collected and has overdue. The tag is optional:
-- existing invoices stay untagged and nothing is backfilled. Fresh databases
-- get the column and index from finance/schema.sql, so both steps below are
-- no-ops there, and a stopped run can be retried.

SET @invoice_project_column_missing := (
    SELECT COUNT(*) = 0
    FROM information_schema.columns
    WHERE table_schema = DATABASE()
      AND table_name = 'finance_invoices'
      AND column_name = 'project_id'
);
SET @invoice_project_sql := IF(
    @invoice_project_column_missing,
    'ALTER TABLE finance_invoices ADD COLUMN project_id bigint DEFAULT NULL AFTER customer_id',
    'SELECT 1'
);
PREPARE invoice_project_stmt FROM @invoice_project_sql;
EXECUTE invoice_project_stmt;
DEALLOCATE PREPARE invoice_project_stmt;

SET @invoice_project_index_missing := (
    SELECT COUNT(*) = 0
    FROM information_schema.statistics
    WHERE table_schema = DATABASE()
      AND table_name = 'finance_invoices'
      AND index_name = 'idx_finance_invoices_business_project'
);
SET @invoice_project_sql := IF(
    @invoice_project_index_missing,
    'ALTER TABLE finance_invoices ADD KEY idx_finance_invoices_business_project (business_id, project_id)',
    'SELECT 1'
);
PREPARE invoice_project_stmt FROM @invoice_project_sql;
EXECUTE invoice_project_stmt;
DEALLOCATE PREPARE invoice_project_stmt;
