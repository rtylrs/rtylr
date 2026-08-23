-- F-B: a person-to-person reporting line on the staff record, feeding Flow's
-- dept-manager visibility proxy (M3) as a second rule source alongside
-- hr_departments.manager_staff_id.
--
-- Nullable, and no default: most staff have no explicit manager, and a NULL
-- reads as "no direct line" rather than "reports to staff 0". Self-reference is
-- prevented in the resolver rather than by a constraint, because the check
-- (manager_staff_id <> id) cannot be expressed as a column DEFAULT and a CHECK
-- would reject legitimate mid-edit states.
--
-- Deliberately NOT a foreign key to hr_staff(id): the rest of this schema does
-- not use FKs (hr_departments.manager_staff_id has none either), and adding one
-- here alone would make staff deletion fail in a way the code does not expect.
--
-- Own filename plus an INFORMATION_SCHEMA guard: migrate.sh records applied
-- files by name so editing hr/schema.sql alone would be a silent no-op on
-- deployed databases, and zz_ files re-run every pass while a FAILED migration
-- is recorded as applied behind a green check.
SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
   WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hr_staff'
     AND COLUMN_NAME = 'manager_staff_id') = 0,
  'ALTER TABLE hr_staff ADD COLUMN manager_staff_id BIGINT NULL AFTER department_id',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @idx := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
   WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hr_staff'
     AND INDEX_NAME = 'idx_hr_staff_manager') = 0,
  'ALTER TABLE hr_staff ADD INDEX idx_hr_staff_manager (business_id, manager_staff_id)',
  'SELECT 1'
);
PREPARE stmt2 FROM @idx; EXECUTE stmt2; DEALLOCATE PREPARE stmt2;
