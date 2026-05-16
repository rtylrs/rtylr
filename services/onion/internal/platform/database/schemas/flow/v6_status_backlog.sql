-- Flow schema migration v6: add 'backlog' to flow_tasks.status ENUM
--
-- The original status ENUM was ('todo','in_progress','in_review','done',
-- 'cancelled'). The Flow UI uses 'backlog' as the leftmost kanban column
-- on My Tasks (so quick-captures land somewhere unscheduled instead of
-- jamming Todo). Without this migration, strict mode rejects the INSERT
-- with "Data truncated for column 'status'" and any quick-add 500s.
--
-- The MODIFY COLUMN preserves existing rows (each row keeps whatever
-- value it already had); we're only widening the set of valid values.

ALTER TABLE flow_tasks
  MODIFY COLUMN status
    ENUM('backlog','todo','in_progress','in_review','done','cancelled')
    DEFAULT 'todo';
