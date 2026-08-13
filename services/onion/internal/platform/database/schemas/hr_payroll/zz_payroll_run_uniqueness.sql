-- A payroll period may be triggered concurrently. This database constraint is
-- the authoritative guard; application checks only provide the friendly fast
-- path. Existing owner-period indexes remain for legacy owner-scoped reads.
SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hr_payroll_runs' AND INDEX_NAME = 'uk_payroll_runs_business_period') = 0,
  'ALTER TABLE hr_payroll_runs ADD UNIQUE KEY uk_payroll_runs_business_period (business_id, period_start, period_end)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
