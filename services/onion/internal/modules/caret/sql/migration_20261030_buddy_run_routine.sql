-- Runs started by a routine remember which one, so the buddy can post the
-- result under the routine's name. New databases get the column from
-- caret_buddy_runs.sql; this adds it where that table already exists.

SET @buddy_run_routine_exists := (
    SELECT COUNT(*) > 0
    FROM information_schema.columns
    WHERE table_schema = DATABASE()
      AND table_name = 'caret_buddy_runs'
      AND column_name = 'routine_id'
);
SET @buddy_run_routine_sql := IF(
    @buddy_run_routine_exists,
    'SELECT 1',
    'ALTER TABLE caret_buddy_runs ADD COLUMN routine_id bigint NULL DEFAULT NULL AFTER buddy_id'
);
PREPARE buddy_run_routine FROM @buddy_run_routine_sql;
EXECUTE buddy_run_routine;
DEALLOCATE PREPARE buddy_run_routine;
