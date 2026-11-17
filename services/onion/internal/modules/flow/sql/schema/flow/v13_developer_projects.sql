-- Developer projects (T-365): one flag that later developer features (code
-- links, CI, dev fields) check. Guarded so a re-run is a no-op.
SET @flow_developer_mode_missing := (
    SELECT COUNT(*) = 0
    FROM information_schema.columns
    WHERE table_schema = DATABASE()
      AND table_name = 'flow_projects'
      AND column_name = 'developer_mode'
);
SET @flow_developer_mode_sql := IF(
    @flow_developer_mode_missing,
    'ALTER TABLE `flow_projects` ADD COLUMN `developer_mode` tinyint(1) NOT NULL DEFAULT ''0'' AFTER `is_personal`',
    'SELECT 1'
);
PREPARE flow_developer_mode_statement FROM @flow_developer_mode_sql;
EXECUTE flow_developer_mode_statement;
DEALLOCATE PREPARE flow_developer_mode_statement;
