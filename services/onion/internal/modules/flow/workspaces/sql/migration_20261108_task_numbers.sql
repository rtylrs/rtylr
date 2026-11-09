-- Task numbers (T-n) are counted per business. They used to be the global row
-- id, which showed every business how many tasks the others had made.
-- Existing tasks keep the number they already show (their id), and each
-- business counts on from its highest one, so no visible number changes.

SET @task_number_column_exists := (
    SELECT COUNT(*) > 0
    FROM information_schema.columns
    WHERE table_schema = DATABASE()
      AND table_name = 'flow_tasks'
      AND column_name = 'number'
);
SET @add_task_number_column_sql := IF(
    @task_number_column_exists,
    'SELECT 1',
    "ALTER TABLE flow_tasks ADD COLUMN number bigint DEFAULT NULL COMMENT 'Shown as T-<number>; counted per business by flow_task_sequences' AFTER business_id"
);
PREPARE add_task_number_column_statement FROM @add_task_number_column_sql;
EXECUTE add_task_number_column_statement;
DEALLOCATE PREPARE add_task_number_column_statement;

UPDATE flow_tasks SET number = id WHERE number IS NULL;

SET @task_number_index_exists := (
    SELECT COUNT(*) > 0
    FROM information_schema.statistics
    WHERE table_schema = DATABASE()
      AND table_name = 'flow_tasks'
      AND index_name = 'uk_flow_tasks_business_number'
);
SET @add_task_number_index_sql := IF(
    @task_number_index_exists,
    'SELECT 1',
    'ALTER TABLE flow_tasks ADD UNIQUE KEY uk_flow_tasks_business_number (business_id, number)'
);
PREPARE add_task_number_index_statement FROM @add_task_number_index_sql;
EXECUTE add_task_number_index_statement;
DEALLOCATE PREPARE add_task_number_index_statement;

CREATE TABLE IF NOT EXISTS flow_task_sequences (
  business_id bigint NOT NULL,
  last_number bigint NOT NULL DEFAULT '0',
  PRIMARY KEY (business_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO flow_task_sequences (business_id, last_number)
SELECT business_id, MAX(number) FROM flow_tasks GROUP BY business_id
ON DUPLICATE KEY UPDATE last_number = GREATEST(flow_task_sequences.last_number, VALUES(last_number));
