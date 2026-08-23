-- S1/D2: collapse same-named labels within a business into one.
--
-- Labels used to be workspace-scoped, so "Bug" in three workspaces was three
-- rows. v9_workspace_removal_prep re-scoped them to one business-wide pool,
-- which made those rows duplicates of each other. The proposal tolerated that
-- indefinitely and deferred the dedupe; this is the dedupe.
--
-- The hard part is that labels are NOT a join table. flow_tasks.labels is a JSON
-- array of label ids (schema.sql: "Array of label IDs"), so merging cannot
-- repoint foreign keys: every affected task's array has to be rewritten, with the
-- merged ids mapped to the survivor and the result de-duplicated in case a task
-- carried two labels that are now one.
--
-- Survivor is MIN(id) per (business_id, name): the oldest row wins, so the id
-- that has been referenced longest is the one that persists.
--
-- Idempotent by construction: once no duplicate names remain the mapping is
-- empty, the UPDATE matches nothing and the DELETE removes nothing. Guarded on
-- INFORMATION_SCHEMA because zz_ files re-run every pass and migrate.sh records
-- a FAILED migration as applied behind a green check.

SET @ok := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'flow_labels'
    AND COLUMN_NAME = 'business_id'
);

DROP TEMPORARY TABLE IF EXISTS flow_label_merge;
CREATE TEMPORARY TABLE flow_label_merge (
  old_id BIGINT PRIMARY KEY,
  keep_id BIGINT NOT NULL
);

SET @sql := IF(@ok > 0,
  'INSERT INTO flow_label_merge (old_id, keep_id)
   SELECT l.id, s.keep_id
   FROM flow_labels l
   JOIN (SELECT business_id, name, MIN(id) AS keep_id
         FROM flow_labels GROUP BY business_id, name) s
     ON s.business_id = l.business_id AND s.name = l.name
   WHERE l.id <> s.keep_id',
  'SELECT 1');
PREPARE st FROM @sql; EXECUTE st; DEALLOCATE PREPARE st;

-- Rewrite every affected task's label array. JSON_ARRAYAGG has no DISTINCT in
-- MySQL 8, so the de-duplication happens in the inner SELECT DISTINCT.
--
-- The duplicate count is hoisted into a variable rather than tested with EXISTS
-- inside the UPDATE: a TEMPORARY table cannot be referenced twice in one
-- statement (MySQL 1137, "Can't reopen table"). It also means a run with nothing
-- to merge skips the UPDATE entirely instead of rewriting every task's array to
-- an equivalent value.
SET @dupes := (SELECT COUNT(*) FROM flow_label_merge);

SET @sql2 := IF(@ok > 0 AND @dupes > 0,
  'UPDATE flow_tasks t
   JOIN (
     SELECT task_id, JSON_ARRAYAGG(lid) AS new_labels
     FROM (
       SELECT DISTINCT t2.id AS task_id, COALESCE(m.keep_id, j.label_id) AS lid
       FROM flow_tasks t2
       JOIN JSON_TABLE(t2.labels, ''$[*]'' COLUMNS (label_id BIGINT PATH ''$'')) j
       LEFT JOIN flow_label_merge m ON m.old_id = j.label_id
       WHERE t2.labels IS NOT NULL AND JSON_LENGTH(t2.labels) > 0
     ) d
     GROUP BY task_id
   ) x ON x.task_id = t.id
   SET t.labels = x.new_labels',
  'SELECT 1');
PREPARE st2 FROM @sql2; EXECUTE st2; DEALLOCATE PREPARE st2;

SET @sql3 := IF(@ok > 0,
  'DELETE FROM flow_labels WHERE id IN (SELECT old_id FROM flow_label_merge)',
  'SELECT 1');
PREPARE st3 FROM @sql3; EXECUTE st3; DEALLOCATE PREPARE st3;

DROP TEMPORARY TABLE IF EXISTS flow_label_merge;
