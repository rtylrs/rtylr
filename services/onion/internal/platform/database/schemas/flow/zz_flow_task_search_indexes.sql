-- Indexed Flow task search support.
--
-- Previously three bare ALTERs with the comment "duplicate-index errors are safe
-- under the migration runner's --force mode". They are not safe: migrate.sh
-- records a FAILED migration as applied behind a green check, and zz_ files
-- re-run on every pass, so this file failed silently on every migration after
-- the first and nothing surfaced it.
--
-- Guarded per statement now: the column on INFORMATION_SCHEMA.COLUMNS, the
-- indexes on INFORMATION_SCHEMA.STATISTICS.
--
-- idx_flow_tasks_workspace_updated leads with workspace_id, so it is additionally
-- guarded on that column existing. Phase 4 drops workspace_id, and an index
-- guard alone would then pass (the index is gone) while the ALTER fails (the
-- column is gone) - the same trap zz_flow_hot_path_indexes.sql had.

SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
   WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'flow_tasks'
     AND COLUMN_NAME = 'archived') = 0,
  'ALTER TABLE flow_tasks ADD COLUMN archived BOOLEAN DEFAULT FALSE',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql2 := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
   WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'flow_tasks'
     AND INDEX_NAME = 'idx_flow_tasks_workspace_updated') = 0
  AND (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
       WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'flow_tasks'
         AND COLUMN_NAME = 'workspace_id') > 0,
  'ALTER TABLE flow_tasks ADD INDEX idx_flow_tasks_workspace_updated (workspace_id, archived, updated_at)',
  'SELECT 1'
);
PREPARE stmt2 FROM @sql2; EXECUTE stmt2; DEALLOCATE PREPARE stmt2;

SET @sql3 := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
   WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'flow_tasks'
     AND INDEX_NAME = 'idx_flow_tasks_search') = 0,
  'ALTER TABLE flow_tasks ADD FULLTEXT INDEX idx_flow_tasks_search (title, description)',
  'SELECT 1'
);
PREPARE stmt3 FROM @sql3; EXECUTE stmt3; DEALLOCATE PREPARE stmt3;
