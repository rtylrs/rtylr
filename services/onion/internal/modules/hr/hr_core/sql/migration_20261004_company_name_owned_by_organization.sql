SET @hr_company_name_column_exists := (
    SELECT COUNT(*)
    FROM INFORMATION_SCHEMA.COLUMNS
    WHERE TABLE_SCHEMA = DATABASE()
      AND TABLE_NAME = 'hr_company_config'
      AND COLUMN_NAME = 'company_name'
);
SET @hr_company_name_migration := IF(
    @hr_company_name_column_exists > 0,
    'ALTER TABLE hr_company_config DROP COLUMN company_name',
    'SELECT 1'
);
PREPARE hr_company_name_migration_stmt FROM @hr_company_name_migration;
EXECUTE hr_company_name_migration_stmt;
DEALLOCATE PREPARE hr_company_name_migration_stmt;
