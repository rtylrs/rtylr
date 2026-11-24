-- Story points (T-381): an optional points estimate on a task, separate from
-- effort minutes, and the scale a project picks its points from. Additive, and
-- each step is guarded so a re-run is a no-op.
SET @flow_task_story_points_missing := (
    SELECT COUNT(*) = 0
    FROM information_schema.columns
    WHERE table_schema = DATABASE()
      AND table_name = 'flow_tasks'
      AND column_name = 'story_points'
);
SET @flow_task_story_points_sql := IF(
    @flow_task_story_points_missing,
    'ALTER TABLE `flow_tasks` ADD COLUMN `story_points` smallint DEFAULT NULL AFTER `effort_minutes`',
    'SELECT 1'
);
PREPARE flow_task_story_points_statement FROM @flow_task_story_points_sql;
EXECUTE flow_task_story_points_statement;
DEALLOCATE PREPARE flow_task_story_points_statement;

SET @flow_project_points_scale_missing := (
    SELECT COUNT(*) = 0
    FROM information_schema.columns
    WHERE table_schema = DATABASE()
      AND table_name = 'flow_projects'
      AND column_name = 'points_scale'
);
SET @flow_project_points_scale_sql := IF(
    @flow_project_points_scale_missing,
    'ALTER TABLE `flow_projects` ADD COLUMN `points_scale` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT ''fibonacci'' AFTER `developer_mode`',
    'SELECT 1'
);
PREPARE flow_project_points_scale_statement FROM @flow_project_points_scale_sql;
EXECUTE flow_project_points_scale_statement;
DEALLOCATE PREPARE flow_project_points_scale_statement;
