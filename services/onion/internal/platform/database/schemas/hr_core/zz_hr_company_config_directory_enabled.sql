-- hr_company_config is declared with CREATE TABLE IF NOT EXISTS, so in any
-- environment where the table already exists that statement is skipped whole
-- and the directory_enabled column added to schema.sql in #33 never lands.
-- Only a fresh database picks it up from the CREATE. This file carries the
-- column as its own ALTER so existing databases get it too.
--
-- migrate.sh re-runs every <domain>/*.sql on each invocation and keeps no
-- applied-migration ledger, so this has to be idempotent on its own. The
-- INFORMATION_SCHEMA guard is what makes re-running it a no-op. It also has to
-- sit at the top level of hr_core/: the apply loop globs "$schema_dir"/*.sql
-- one level deep, so anything under hr_core/migrations/ is never run.
SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hr_company_config' AND COLUMN_NAME = 'directory_enabled') = 0,
  'ALTER TABLE hr_company_config ADD COLUMN directory_enabled BOOLEAN NOT NULL DEFAULT FALSE AFTER overtime_multiplier',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
