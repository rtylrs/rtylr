-- Board columns become the project's statuses.
--
-- Until now a task carried a built-in status, an optional custom status_id and
-- a section_id pointing at a board column, and the three could disagree. From
-- here a project's flow_statuses rows are its columns, section_id mirrors
-- status_id, and flow_project_columns is no longer read (kept for rollback).
-- Every statement is idempotent so a stopped run can be retried, and task
-- updated_at is preserved so recently-edited lists keep their order.

CREATE TABLE IF NOT EXISTS `flow_status_templates` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `business_id` bigint NOT NULL,
  `name` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `bucket` enum('backlog','active','done','cancelled') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `legacy_status` enum('backlog','todo','in_progress','in_review','done','cancelled') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `color` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '#6366f1',
  `sort_order` int NOT NULL DEFAULT '0',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_flow_status_templates_business_name` (`business_id`,`name`),
  KEY `idx_flow_status_templates_business` (`business_id`,`sort_order`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Which status each old column became. is_custom marks a column whose name is
-- not one of the built-in states ("Design", "Phase 1"): its unfinished tasks
-- keep their place on the board instead of following their old status.
CREATE TABLE IF NOT EXISTS `flow_status_column_map` (
  `column_id` bigint NOT NULL,
  `project_id` bigint NOT NULL,
  `status_id` bigint NOT NULL,
  `is_custom` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`column_id`),
  KEY `idx_flow_status_column_map_status` (`status_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 1. A status for every column whose name the project does not already have.
INSERT IGNORE INTO `flow_statuses`
  (`business_id`, `project_id`, `name`, `bucket`, `legacy_status`, `color`, `sort_order`, `is_default`, `created_by`)
SELECT p.`business_id`, c.`project_id`, LEFT(TRIM(c.`name`), 64),
  CASE LOWER(TRIM(c.`name`))
    WHEN 'backlog' THEN 'backlog'
    WHEN 'to do' THEN 'backlog' WHEN 'todo' THEN 'backlog' WHEN 'to-do' THEN 'backlog'
    WHEN 'done' THEN 'done' WHEN 'completed' THEN 'done' WHEN 'complete' THEN 'done'
    WHEN 'cancelled' THEN 'cancelled' WHEN 'canceled' THEN 'cancelled'
    ELSE 'active' END,
  CASE LOWER(TRIM(c.`name`))
    WHEN 'backlog' THEN 'backlog'
    WHEN 'to do' THEN 'todo' WHEN 'todo' THEN 'todo' WHEN 'to-do' THEN 'todo'
    WHEN 'in review' THEN 'in_review' WHEN 'review' THEN 'in_review'
    WHEN 'done' THEN 'done' WHEN 'completed' THEN 'done' WHEN 'complete' THEN 'done'
    WHEN 'cancelled' THEN 'cancelled' WHEN 'canceled' THEN 'cancelled'
    ELSE 'in_progress' END,
  COALESCE(NULLIF(c.`color`, ''), '#6366f1'),
  (COALESCE(c.`sort_order`, 0) + 1) * 10,
  FALSE,
  c.`created_by`
FROM `flow_project_columns` c
JOIN `flow_projects` p ON p.`id` = c.`project_id`
WHERE TRIM(c.`name`) <> ''
  AND NOT EXISTS (
    SELECT 1 FROM `flow_statuses` s
    WHERE s.`project_id` = c.`project_id` AND LOWER(TRIM(s.`name`)) = LOWER(TRIM(c.`name`))
  );

-- 2. Remember which status each column became.
INSERT IGNORE INTO `flow_status_column_map` (`column_id`, `project_id`, `status_id`, `is_custom`)
SELECT c.`id`, c.`project_id`, s.`id`,
  LOWER(TRIM(c.`name`)) NOT IN ('backlog', 'to do', 'todo', 'to-do', 'in progress', 'doing', 'in review', 'review',
                                'done', 'completed', 'complete', 'cancelled', 'canceled')
FROM `flow_project_columns` c
JOIN `flow_statuses` s ON s.`project_id` = c.`project_id` AND LOWER(TRIM(s.`name`)) = LOWER(TRIM(c.`name`));

-- 3. Every project gets the six built-in states it does not have yet.
INSERT IGNORE INTO `flow_statuses`
  (`business_id`, `project_id`, `name`, `bucket`, `legacy_status`, `color`, `sort_order`, `is_default`, `created_by`)
SELECT p.`business_id`, p.`id`, d.`name`, d.`bucket`, d.`legacy_status`, d.`color`, d.`sort_order`, FALSE, p.`created_by`
FROM `flow_projects` p
CROSS JOIN (
  SELECT 'Backlog' AS `name`, 'backlog' AS `bucket`, 'backlog' AS `legacy_status`, '#94a3b8' AS `color`, 0 AS `sort_order`
  UNION ALL SELECT 'To Do', 'backlog', 'todo', '#9ca3af', 10
  UNION ALL SELECT 'In Progress', 'active', 'in_progress', '#3b82f6', 20
  UNION ALL SELECT 'In Review', 'active', 'in_review', '#f59e0b', 30
  UNION ALL SELECT 'Done', 'done', 'done', '#22c55e', 900
  UNION ALL SELECT 'Cancelled', 'cancelled', 'cancelled', '#ef4444', 1000
) d
WHERE NOT EXISTS (
  SELECT 1 FROM `flow_statuses` s
  WHERE s.`project_id` = p.`id` AND LOWER(s.`name`) = LOWER(d.`name`)
);

-- 4. New tasks land in the project's To Do unless it already chose a default.
UPDATE `flow_statuses` s
JOIN (
  SELECT s2.`project_id`, MIN(s2.`id`) AS `id`
  FROM `flow_statuses` s2
  WHERE s2.`legacy_status` = 'todo'
    AND NOT EXISTS (SELECT 1 FROM `flow_statuses` s3 WHERE s3.`project_id` = s2.`project_id` AND s3.`is_default` = TRUE)
  GROUP BY s2.`project_id`
) pick ON pick.`id` = s.`id`
SET s.`is_default` = TRUE;

-- 5. Unfinished tasks sitting in a custom column stay in it.
UPDATE `flow_tasks` t
JOIN `flow_status_column_map` m ON m.`column_id` = t.`section_id` AND m.`project_id` = t.`project_id` AND m.`is_custom` = 1
JOIN `flow_statuses` s ON s.`id` = m.`status_id`
SET t.`status_id` = s.`id`, t.`status` = s.`legacy_status`, t.`updated_at` = t.`updated_at`
WHERE t.`status` NOT IN ('done', 'cancelled')
  AND (t.`status_id` IS NULL OR t.`status_id` = s.`id`);

-- 6. Everything else: the task's status wins. A custom status_id that already
-- belongs to the task's project and agrees with its built-in status is kept.
UPDATE `flow_tasks` t
SET t.`status_id` = COALESCE(
  (SELECT s.`id` FROM `flow_statuses` s
   WHERE s.`id` = t.`status_id` AND s.`project_id` = t.`project_id` AND s.`legacy_status` = t.`status`),
  (SELECT s.`id` FROM `flow_statuses` s
   WHERE s.`project_id` = t.`project_id` AND s.`legacy_status` = t.`status`
   ORDER BY (LOWER(s.`name`) IN ('backlog', 'to do', 'in progress', 'in review', 'done', 'cancelled')) DESC,
            s.`sort_order`, s.`id`
   LIMIT 1)
), t.`updated_at` = t.`updated_at`;

-- 7. section_id mirrors status_id from now on.
UPDATE `flow_tasks` SET `section_id` = `status_id`, `updated_at` = `updated_at`
WHERE NOT (`section_id` <=> `status_id`);
