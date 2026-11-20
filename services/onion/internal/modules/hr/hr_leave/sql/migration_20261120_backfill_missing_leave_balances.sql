-- Leave balances were only created when an employee was added through the HR
-- employee form. Employees provisioned from a business membership, and anyone
-- hired before a leave type existed, had no balance row and every leave
-- request failed with "Leave balance is not configured for this employee".
--
-- Gives each non-terminated employee a current-year balance for every active
-- leave type they lack. A new row starts with what the monthly accrual job's
-- catch-up would grant on its next run (the full annual amount, or one monthly
-- amount per month elapsed), so the balance is usable straight away. Existing
-- rows are never touched, which makes the statement safe to re-run.
INSERT IGNORE INTO hr_leave_balances
    (owner_id, business_id, employee_id, leave_type_id, year, entitled_days, last_accrual_month)
SELECT employee.business_id, employee.business_id, employee.id, leave_type.id, YEAR(UTC_DATE()),
       CASE
           WHEN COALESCE(leave_type.accrual_amount, 0) <= 0 THEN 0
           WHEN leave_type.accrual_rule = 'annual' THEN leave_type.accrual_amount
           WHEN leave_type.accrual_rule = 'monthly' THEN leave_type.accrual_amount * MONTH(UTC_DATE())
           ELSE 0
       END,
       CASE
           WHEN COALESCE(leave_type.accrual_amount, 0) <= 0 THEN 0
           WHEN leave_type.accrual_rule = 'annual' THEN 12
           WHEN leave_type.accrual_rule = 'monthly' THEN MONTH(UTC_DATE())
           ELSE 0
       END
FROM hr_employees AS employee
JOIN hr_leave_types AS leave_type
  ON leave_type.business_id = employee.business_id
 AND leave_type.is_active = TRUE
WHERE employee.business_id > 0
  AND employee.status <> 'terminated'
  AND NOT EXISTS (
      SELECT 1
      FROM hr_leave_balances AS existing
      WHERE existing.business_id = employee.business_id
        AND existing.employee_id = employee.id
        AND existing.leave_type_id = leave_type.id
        AND existing.year = YEAR(UTC_DATE())
  );
