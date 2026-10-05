-- CREATE TABLE IF NOT EXISTS does not upgrade settings tables created before
-- the reporting preferences were added to the canonical Insights schema.
SET @insights_settings_sql = IF(
    EXISTS (SELECT 1 FROM information_schema.columns
            WHERE table_schema = DATABASE() AND table_name = 'insights_settings'
              AND column_name = 'default_period'),
    'SELECT 1',
    'ALTER TABLE insights_settings ADD COLUMN default_period VARCHAR(20) NOT NULL DEFAULT ''this_month'''
);
PREPARE insights_settings_stmt FROM @insights_settings_sql;
EXECUTE insights_settings_stmt;
DEALLOCATE PREPARE insights_settings_stmt;

SET @insights_settings_sql = IF(
    EXISTS (SELECT 1 FROM information_schema.columns
            WHERE table_schema = DATABASE() AND table_name = 'insights_settings'
              AND column_name = 'compare_previous'),
    'SELECT 1',
    'ALTER TABLE insights_settings ADD COLUMN compare_previous TINYINT(1) NOT NULL DEFAULT 0'
);
PREPARE insights_settings_stmt FROM @insights_settings_sql;
EXECUTE insights_settings_stmt;
DEALLOCATE PREPARE insights_settings_stmt;

SET @insights_settings_sql = IF(
    EXISTS (SELECT 1 FROM information_schema.columns
            WHERE table_schema = DATABASE() AND table_name = 'insights_settings'
              AND column_name = 'dashboard_widgets'),
    'SELECT 1',
    'ALTER TABLE insights_settings ADD COLUMN dashboard_widgets JSON NOT NULL DEFAULT (JSON_ARRAY())'
);
PREPARE insights_settings_stmt FROM @insights_settings_sql;
EXECUTE insights_settings_stmt;
DEALLOCATE PREPARE insights_settings_stmt;

SET @insights_settings_sql = IF(
    EXISTS (SELECT 1 FROM information_schema.columns
            WHERE table_schema = DATABASE() AND table_name = 'insights_settings'
              AND column_name = 'favorite_reports'),
    'SELECT 1',
    'ALTER TABLE insights_settings ADD COLUMN favorite_reports JSON NOT NULL DEFAULT (JSON_ARRAY())'
);
PREPARE insights_settings_stmt FROM @insights_settings_sql;
EXECUTE insights_settings_stmt;
DEALLOCATE PREPARE insights_settings_stmt;
