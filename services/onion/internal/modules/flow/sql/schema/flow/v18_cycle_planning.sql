-- Cycle planning (T-336, T-571): a goal and lifecycle stamps on a cycle, a rank
-- for a task inside a cycle and inside the unplanned backlog, and a record of
-- what each unfinished task did when its cycle closed. Additive, and every
-- statement is guarded so a re-run is a no-op.
SET @flow_cycle_goal_missing := (
    SELECT COUNT(*) = 0
    FROM information_schema.columns
    WHERE table_schema = DATABASE()
      AND table_name = 'flow_cycles'
      AND column_name = 'goal'
);
SET @flow_cycle_goal_sql := IF(
    @flow_cycle_goal_missing,
    'ALTER TABLE `flow_cycles` ADD COLUMN `goal` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci AFTER `status`',
    'SELECT 1'
);
PREPARE flow_cycle_goal_statement FROM @flow_cycle_goal_sql;
EXECUTE flow_cycle_goal_statement;
DEALLOCATE PREPARE flow_cycle_goal_statement;

SET @flow_cycle_started_at_missing := (
    SELECT COUNT(*) = 0
    FROM information_schema.columns
    WHERE table_schema = DATABASE()
      AND table_name = 'flow_cycles'
      AND column_name = 'started_at'
);
SET @flow_cycle_started_at_sql := IF(
    @flow_cycle_started_at_missing,
    'ALTER TABLE `flow_cycles` ADD COLUMN `started_at` timestamp NULL DEFAULT NULL AFTER `goal`',
    'SELECT 1'
);
PREPARE flow_cycle_started_at_statement FROM @flow_cycle_started_at_sql;
EXECUTE flow_cycle_started_at_statement;
DEALLOCATE PREPARE flow_cycle_started_at_statement;

SET @flow_cycle_completed_at_missing := (
    SELECT COUNT(*) = 0
    FROM information_schema.columns
    WHERE table_schema = DATABASE()
      AND table_name = 'flow_cycles'
      AND column_name = 'completed_at'
);
SET @flow_cycle_completed_at_sql := IF(
    @flow_cycle_completed_at_missing,
    'ALTER TABLE `flow_cycles` ADD COLUMN `completed_at` timestamp NULL DEFAULT NULL AFTER `started_at`',
    'SELECT 1'
);
PREPARE flow_cycle_completed_at_statement FROM @flow_cycle_completed_at_sql;
EXECUTE flow_cycle_completed_at_statement;
DEALLOCATE PREPARE flow_cycle_completed_at_statement;

SET @flow_cycle_project_idx_missing := (
    SELECT COUNT(*) = 0
    FROM information_schema.statistics
    WHERE table_schema = DATABASE()
      AND table_name = 'flow_cycles'
      AND index_name = 'idx_flow_cycles_business_project'
);
SET @flow_cycle_project_idx_sql := IF(
    @flow_cycle_project_idx_missing,
    'ALTER TABLE `flow_cycles` ADD KEY `idx_flow_cycles_business_project` (`business_id`,`project_id`,`status`)',
    'SELECT 1'
);
PREPARE flow_cycle_project_idx_statement FROM @flow_cycle_project_idx_sql;
EXECUTE flow_cycle_project_idx_statement;
DEALLOCATE PREPARE flow_cycle_project_idx_statement;

SET @flow_cycle_task_rank_missing := (
    SELECT COUNT(*) = 0
    FROM information_schema.columns
    WHERE table_schema = DATABASE()
      AND table_name = 'flow_cycle_tasks'
      AND column_name = 'sort_rank'
);
SET @flow_cycle_task_rank_sql := IF(
    @flow_cycle_task_rank_missing,
    'ALTER TABLE `flow_cycle_tasks` ADD COLUMN `sort_rank` bigint NOT NULL DEFAULT ''0'' AFTER `added_at`',
    'SELECT 1'
);
PREPARE flow_cycle_task_rank_statement FROM @flow_cycle_task_rank_sql;
EXECUTE flow_cycle_task_rank_statement;
DEALLOCATE PREPARE flow_cycle_task_rank_statement;

SET @flow_cycle_task_idx_missing := (
    SELECT COUNT(*) = 0
    FROM information_schema.statistics
    WHERE table_schema = DATABASE()
      AND table_name = 'flow_cycle_tasks'
      AND index_name = 'idx_flow_cycle_tasks_task'
);
SET @flow_cycle_task_idx_sql := IF(
    @flow_cycle_task_idx_missing,
    'ALTER TABLE `flow_cycle_tasks` ADD KEY `idx_flow_cycle_tasks_task` (`task_id`)',
    'SELECT 1'
);
PREPARE flow_cycle_task_idx_statement FROM @flow_cycle_task_idx_sql;
EXECUTE flow_cycle_task_idx_statement;
DEALLOCATE PREPARE flow_cycle_task_idx_statement;

SET @flow_task_backlog_rank_missing := (
    SELECT COUNT(*) = 0
    FROM information_schema.columns
    WHERE table_schema = DATABASE()
      AND table_name = 'flow_tasks'
      AND column_name = 'backlog_rank'
);
SET @flow_task_backlog_rank_sql := IF(
    @flow_task_backlog_rank_missing,
    'ALTER TABLE `flow_tasks` ADD COLUMN `backlog_rank` bigint DEFAULT NULL AFTER `sort_order`',
    'SELECT 1'
);
PREPARE flow_task_backlog_rank_statement FROM @flow_task_backlog_rank_sql;
EXECUTE flow_task_backlog_rank_statement;
DEALLOCATE PREPARE flow_task_backlog_rank_statement;

CREATE TABLE IF NOT EXISTS `flow_cycle_carryovers` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `business_id` bigint NOT NULL,
  `cycle_id` bigint NOT NULL,
  `task_id` bigint NOT NULL,
  `to_cycle_id` bigint DEFAULT NULL,
  `decision` enum('move','backlog','cancel') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `created_by` bigint NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_flow_cycle_carryovers_cycle` (`business_id`,`cycle_id`),
  KEY `idx_flow_cycle_carryovers_to_cycle` (`business_id`,`to_cycle_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
