SET @mfa_enforcement_column_missing := (
    SELECT COUNT(*) = 0
    FROM information_schema.columns
    WHERE table_schema = DATABASE()
      AND table_name = 'businesses'
      AND column_name = 'mfa_enforced'
);
SET @mfa_enforcement_sql := IF(
    @mfa_enforcement_column_missing,
    'ALTER TABLE `businesses` ADD COLUMN `mfa_enforced` tinyint(1) NOT NULL DEFAULT ''0'' AFTER `is_active`',
    'SELECT 1'
);
PREPARE mfa_enforcement_statement FROM @mfa_enforcement_sql;
EXECUTE mfa_enforcement_statement;
DEALLOCATE PREPARE mfa_enforcement_statement;
