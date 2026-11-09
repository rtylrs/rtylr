-- Ties each project's four default board columns to the status they stand
-- for, so dropping a card on Done marks it done and marking a task done moves
-- its card to Done. Columns people added stay unlinked. Only the oldest column
-- of each default name in a project is linked, because some projects were
-- seeded twice, and a renamed default column is left unlinked. No task moves.

SET @column_status_exists := (
    SELECT COUNT(*) > 0
    FROM information_schema.columns
    WHERE table_schema = DATABASE()
      AND table_name = 'flow_project_columns'
      AND column_name = 'linked_status'
);
SET @add_column_status_sql := IF(
    @column_status_exists,
    'SELECT 1',
    "ALTER TABLE flow_project_columns ADD COLUMN linked_status enum('backlog','todo','in_progress','in_review','done','cancelled') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'Status a card takes when dropped here, and the column a task moves to when it takes this status' AFTER sort_order"
);
PREPARE add_column_status_statement FROM @add_column_status_sql;
EXECUTE add_column_status_statement;
DEALLOCATE PREPARE add_column_status_statement;

UPDATE flow_project_columns c
JOIN (
    SELECT first_id FROM (
        SELECT MIN(id) AS first_id
        FROM flow_project_columns
        WHERE LOWER(TRIM(name)) IN ('to do', 'in progress', 'in review', 'done')
        GROUP BY project_id, LOWER(TRIM(name))
    ) AS firsts
) AS f ON f.first_id = c.id
SET c.linked_status = CASE LOWER(TRIM(c.name))
        WHEN 'to do' THEN 'todo'
        WHEN 'in progress' THEN 'in_progress'
        WHEN 'in review' THEN 'in_review'
        WHEN 'done' THEN 'done'
    END,
    c.updated_at = c.updated_at
WHERE c.linked_status IS NULL
  AND NOT EXISTS (
      SELECT 1 FROM (
          SELECT project_id, linked_status FROM flow_project_columns
          WHERE linked_status IS NOT NULL
          GROUP BY project_id, linked_status
      ) AS linked
      WHERE linked.project_id = c.project_id
        AND linked.linked_status = CASE LOWER(TRIM(c.name))
            WHEN 'to do' THEN 'todo'
            WHEN 'in progress' THEN 'in_progress'
            WHEN 'in review' THEN 'in_review'
            WHEN 'done' THEN 'done'
        END
  );

SET @column_status_index_exists := (
    SELECT COUNT(*) > 0
    FROM information_schema.statistics
    WHERE table_schema = DATABASE()
      AND table_name = 'flow_project_columns'
      AND index_name = 'uk_flow_columns_project_status'
);
SET @add_column_status_index_sql := IF(
    @column_status_index_exists,
    'SELECT 1',
    'ALTER TABLE flow_project_columns ADD UNIQUE KEY uk_flow_columns_project_status (project_id, linked_status)'
);
PREPARE add_column_status_index_statement FROM @add_column_status_index_sql;
EXECUTE add_column_status_index_statement;
DEALLOCATE PREPARE add_column_status_index_statement;
