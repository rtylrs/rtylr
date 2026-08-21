-- flow_tasks is declared with CREATE TABLE IF NOT EXISTS, so in any environment
-- where the table already exists that statement is skipped whole and the three
-- sidebar columns added to flow/schema.sql never land. Only a fresh database
-- picks them up from the CREATE. This file carries them as their own ALTERs so
-- existing databases get them too.
--
-- migrate.sh re-runs every <domain>/*.sql on each invocation and keeps no
-- applied-migration ledger, so each statement has to be idempotent on its own.
-- The INFORMATION_SCHEMA guard is what makes re-running this a no-op. It also
-- has to sit at the top level of flow/: the apply loop walks a fixed list of
-- domain directories and globs "$schema_dir"/*.sql one level deep, so a
-- top-level migrations/ directory is never visited at all.
--
-- flow_task_focus_pins and flow_task_dependencies need no equivalent here:
-- their CREATE TABLE IF NOT EXISTS in flow/schema.sql does run against an
-- existing database, because those tables do not exist yet.
SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'flow_tasks' AND COLUMN_NAME = 'effort_minutes') = 0,
  'ALTER TABLE flow_tasks ADD COLUMN effort_minutes INT DEFAULT NULL COMMENT ''Quick effort estimate in minutes, distinct from estimated_hours'' AFTER actual_hours',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'flow_tasks' AND COLUMN_NAME = 'waiting_on') = 0,
  'ALTER TABLE flow_tasks ADD COLUMN waiting_on VARCHAR(255) DEFAULT NULL COMMENT ''Free-text blocker: person, team or approval'' AFTER effort_minutes',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'flow_tasks' AND COLUMN_NAME = 'decision_note') = 0,
  'ALTER TABLE flow_tasks ADD COLUMN decision_note TEXT DEFAULT NULL COMMENT ''One-line decision log for this task'' AFTER waiting_on',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
