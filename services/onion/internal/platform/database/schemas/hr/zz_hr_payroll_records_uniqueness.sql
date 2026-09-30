-- RunPayroll is a POST with no caller-supplied idempotency key, so a double
-- click, a client retry or two managers running the same period each insert
-- their own copy of every pay record. This key is the authoritative guard
-- against the concurrent case; the repository replaces a period's records
-- inside one transaction for the sequential re-run case.
--
-- period_label is the third column because it is the value a re-run repeats and
-- the value ListPayrollRecordsByPeriod filters on, so one run of a period owns
-- exactly one row per staff member.
--
-- Reconciling duplicates already in a deployed database is an operational
-- decision - which row survived, and whether it was already paid - so this
-- migration fails closed instead of deleting or rewriting payroll rows. It
-- prints the offending groups, then aborts before the ALTER. sqlc parses
-- migration files as semicolon-delimited statements, so the abort uses a
-- conditional invalid statement rather than a stored procedure.
--
-- To reconcile, inspect each printed group, keep the row that was actually
-- paid, and delete the rest before re-running:
--
--   SELECT id, staff_id, period_label, status, paid_at, net_pay
--   FROM hr_payroll_records
--   WHERE business_id = ? AND staff_id = ? AND period_label = ?
--   ORDER BY id;
SET @payroll_duplicate_groups := (
    SELECT COUNT(*)
    FROM (
        SELECT business_id, staff_id, period_label
        FROM hr_payroll_records
        GROUP BY business_id, staff_id, period_label
        HAVING COUNT(*) > 1
    ) duplicate_rows
);

SET @sql := IF(
    @payroll_duplicate_groups > 0,
    'SELECT business_id, staff_id, period_label, COUNT(*) AS duplicate_rows, SUM(net_pay) AS duplicated_net_pay FROM hr_payroll_records GROUP BY business_id, staff_id, period_label HAVING COUNT(*) > 1',
    'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
    @payroll_duplicate_groups > 0,
    'SELECT hr_payroll_records_duplicate_periods_must_be_reconciled_before_adding_uq_hr_payroll_records_business_staff_period_label',
    'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
    (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
     WHERE TABLE_SCHEMA = DATABASE()
       AND TABLE_NAME = 'hr_payroll_records'
       AND INDEX_NAME = 'uq_hr_payroll_records_business_staff_period_label') = 0,
    'ALTER TABLE hr_payroll_records ADD UNIQUE KEY uq_hr_payroll_records_business_staff_period_label (business_id, staff_id, period_label)',
    'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
