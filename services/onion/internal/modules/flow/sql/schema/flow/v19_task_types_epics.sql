-- Task types and epics (T-380, T-574). A task is a task, bug, feature or epic.
-- An epic's children point at it through epic_id, so parent_task_id keeps
-- meaning "subtask of" and every existing top-level read stays as it was.
-- Existing rows take the 'task' default; no data is rewritten. Each step is
-- guarded so a re-run is a no-op.
SET @flow_task_type_missing := (
    SELECT COUNT(*) = 0
    FROM information_schema.columns
    WHERE table_schema = DATABASE()
      AND table_name = 'flow_tasks'
      AND column_name = 'task_type'
);
SET @flow_task_type_sql := IF(
    @flow_task_type_missing,
    'ALTER TABLE `flow_tasks` ADD COLUMN `task_type` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT ''task'' AFTER `labels`',
    'SELECT 1'
);
PREPARE flow_task_type_statement FROM @flow_task_type_sql;
EXECUTE flow_task_type_statement;
DEALLOCATE PREPARE flow_task_type_statement;

SET @flow_task_epic_id_missing := (
    SELECT COUNT(*) = 0
    FROM information_schema.columns
    WHERE table_schema = DATABASE()
      AND table_name = 'flow_tasks'
      AND column_name = 'epic_id'
);
SET @flow_task_epic_id_sql := IF(
    @flow_task_epic_id_missing,
    'ALTER TABLE `flow_tasks` ADD COLUMN `epic_id` bigint DEFAULT NULL AFTER `task_type`',
    'SELECT 1'
);
PREPARE flow_task_epic_id_statement FROM @flow_task_epic_id_sql;
EXECUTE flow_task_epic_id_statement;
DEALLOCATE PREPARE flow_task_epic_id_statement;

SET @flow_task_epic_color_missing := (
    SELECT COUNT(*) = 0
    FROM information_schema.columns
    WHERE table_schema = DATABASE()
      AND table_name = 'flow_tasks'
      AND column_name = 'epic_color'
);
SET @flow_task_epic_color_sql := IF(
    @flow_task_epic_color_missing,
    'ALTER TABLE `flow_tasks` ADD COLUMN `epic_color` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL AFTER `epic_id`',
    'SELECT 1'
);
PREPARE flow_task_epic_color_statement FROM @flow_task_epic_color_sql;
EXECUTE flow_task_epic_color_statement;
DEALLOCATE PREPARE flow_task_epic_color_statement;

SET @flow_task_epic_index_missing := (
    SELECT COUNT(*) = 0
    FROM information_schema.statistics
    WHERE table_schema = DATABASE()
      AND table_name = 'flow_tasks'
      AND index_name = 'idx_flow_tasks_epic'
);
SET @flow_task_epic_index_sql := IF(
    @flow_task_epic_index_missing,
    'ALTER TABLE `flow_tasks` ADD INDEX `idx_flow_tasks_epic` (`epic_id`)',
    'SELECT 1'
);
PREPARE flow_task_epic_index_statement FROM @flow_task_epic_index_sql;
EXECUTE flow_task_epic_index_statement;
DEALLOCATE PREPARE flow_task_epic_index_statement;

SET @flow_task_type_index_missing := (
    SELECT COUNT(*) = 0
    FROM information_schema.statistics
    WHERE table_schema = DATABASE()
      AND table_name = 'flow_tasks'
      AND index_name = 'idx_flow_tasks_project_type'
);
SET @flow_task_type_index_sql := IF(
    @flow_task_type_index_missing,
    'ALTER TABLE `flow_tasks` ADD INDEX `idx_flow_tasks_project_type` (`project_id`, `task_type`)',
    'SELECT 1'
);
PREPARE flow_task_type_index_statement FROM @flow_task_type_index_sql;
EXECUTE flow_task_type_index_statement;
DEALLOCATE PREPARE flow_task_type_index_statement;
