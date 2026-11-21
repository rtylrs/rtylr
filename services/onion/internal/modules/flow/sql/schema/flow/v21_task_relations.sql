-- Typed task relations (T-585). flow_task_dependencies gains link_type:
-- blocks, relates, duplicates or clones. Existing rows mean "task_id is
-- blocked by depends_on_task_id", which the column default keeps true with no
-- data rewrite. Every step is guarded so a re-run is a no-op.
SET @flow_task_link_type_missing := (
    SELECT COUNT(*) = 0
    FROM information_schema.columns
    WHERE table_schema = DATABASE()
      AND table_name = 'flow_task_dependencies'
      AND column_name = 'link_type'
);
SET @flow_task_link_type_sql := IF(
    @flow_task_link_type_missing,
    'ALTER TABLE `flow_task_dependencies` ADD COLUMN `link_type` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT ''blocks'' AFTER `depends_on_task_id`',
    'SELECT 1'
);
PREPARE flow_task_link_type_statement FROM @flow_task_link_type_sql;
EXECUTE flow_task_link_type_statement;
DEALLOCATE PREPARE flow_task_link_type_statement;

-- The unique key has to include link_type so one pair of tasks can hold more
-- than one relation. Dropped only while it still lacks the column, then added
-- only while absent, so a run stopped between the two steps resumes cleanly.
SET @flow_task_link_key_stale := (
    SELECT COUNT(*) > 0
    FROM information_schema.statistics
    WHERE table_schema = DATABASE()
      AND table_name = 'flow_task_dependencies'
      AND index_name = 'uq_flow_task_dependencies'
) AND (
    SELECT COUNT(*) = 0
    FROM information_schema.statistics
    WHERE table_schema = DATABASE()
      AND table_name = 'flow_task_dependencies'
      AND index_name = 'uq_flow_task_dependencies'
      AND column_name = 'link_type'
);
SET @flow_task_link_key_drop_sql := IF(
    @flow_task_link_key_stale,
    'ALTER TABLE `flow_task_dependencies` DROP INDEX `uq_flow_task_dependencies`',
    'SELECT 1'
);
PREPARE flow_task_link_key_drop_statement FROM @flow_task_link_key_drop_sql;
EXECUTE flow_task_link_key_drop_statement;
DEALLOCATE PREPARE flow_task_link_key_drop_statement;

SET @flow_task_link_key_missing := (
    SELECT COUNT(*) = 0
    FROM information_schema.statistics
    WHERE table_schema = DATABASE()
      AND table_name = 'flow_task_dependencies'
      AND index_name = 'uq_flow_task_dependencies'
);
SET @flow_task_link_key_add_sql := IF(
    @flow_task_link_key_missing,
    'ALTER TABLE `flow_task_dependencies` ADD UNIQUE KEY `uq_flow_task_dependencies` (`task_id`,`depends_on_task_id`,`link_type`)',
    'SELECT 1'
);
PREPARE flow_task_link_key_add_statement FROM @flow_task_link_key_add_sql;
EXECUTE flow_task_link_key_add_statement;
DEALLOCATE PREPARE flow_task_link_key_add_statement;
