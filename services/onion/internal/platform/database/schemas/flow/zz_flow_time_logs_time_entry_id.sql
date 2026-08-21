-- Overlap key between Flow's two stores of hours.
--
-- Flow writes hours in two places, both from shipped UI. Work-day lines go to
-- hr_time_entries carrying a task_id; the task detail time panel writes
-- flow_time_logs. Nothing linked them, so the per-task rollups had to pick one
-- store and did: RecalcActualHours and GetInsightsTimeByMember summed
-- flow_time_logs alone, and every work-day line counted as zero against its own
-- task. That is tryrtylr/rtylr#903.
--
-- The fix sums both. Summing both needs a way to say "this ad-hoc log and that
-- work-day line are the same hour", or a user who records work in both places
-- is counted twice. time_entry_id is that key: when set, this row duplicates the
-- named hr_time_entries row and the rollups skip it, counting the hr_time_entries
-- side as authoritative because payroll reads it.
--
-- NULL means "not known to duplicate anything", which is every existing row.
-- Nothing writes the column yet, so the immediate behaviour is a plain sum of
-- both stores; linking the two at write time is a UI decision recorded on #903.
--
-- No foreign key on purpose. hr_time_entries is HR's table and flow_time_logs is
-- Flow's, and a cross-module FK would make deleting a time entry fail rather
-- than orphan a pointer the rollups already treat as "count this row".
--
-- Its own filename because migrate.sh records schema files as applied by
-- filename: editing flow/schema.sql would silently no-op in every environment
-- that already applied it.
SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'flow_time_logs' AND COLUMN_NAME = 'time_entry_id') = 0,
  'ALTER TABLE flow_time_logs ADD COLUMN time_entry_id BIGINT NULL AFTER task_id',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- The rollups filter on it, so it needs to be indexed alongside task_id.
SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'flow_time_logs' AND INDEX_NAME = 'idx_flow_time_logs_task_entry') = 0,
  'ALTER TABLE flow_time_logs ADD INDEX idx_flow_time_logs_task_entry (task_id, time_entry_id)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
