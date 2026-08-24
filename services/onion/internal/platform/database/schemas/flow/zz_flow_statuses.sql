-- F-A: per-project custom statuses with semantic buckets.
--
-- Additive by construction. flow_statuses starts empty, flow_tasks.status_id
-- starts NULL on every row, and flow_tasks.status keeps its six-value ENUM and
-- stays populated. A project with no rows in flow_statuses behaves exactly as
-- it does today, so this file changes no behaviour on its own.
--
-- Guarded on INFORMATION_SCHEMA throughout because zz_ files re-run on every
-- pass and migrate.sh records a FAILED migration as applied behind a green
-- check, so an unguarded statement would fail silently and forever after the
-- first run. Named as its own zz_ file rather than an edit to an applied one,
-- because migrate.sh records applied files by name and editing one is a no-op.

-- CREATE TABLE IF NOT EXISTS is already idempotent, so it needs no wrapper.
-- Kept here as well as in flow/schema.sql on purpose: schema.sql is what sqlc
-- reads, this file is what a live database gets. schemas#59 added a migration
-- without the declaration and the column was invisible to sqlc; schemas#63 then
-- made the inverse mistake. Both halves, every time.
CREATE TABLE IF NOT EXISTS flow_statuses (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    -- No DEFAULT 0, deliberately unlike the eleven sibling flow_* tables. That
    -- default is what made the task-dependency tenancy bug silent: rows written
    -- with business_id 0 and read with the caller's real id, so the write
    -- succeeded and the row was never visible. Here a missing write errors.
    business_id BIGINT NOT NULL,
    project_id BIGINT NOT NULL,
    name VARCHAR(64) NOT NULL,
    bucket ENUM('backlog','active','done','cancelled') NOT NULL,
    legacy_status ENUM('backlog','todo','in_progress','in_review','done','cancelled') NOT NULL,
    color VARCHAR(20) NOT NULL DEFAULT '#6366f1',
    sort_order INT NOT NULL DEFAULT 0,
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    created_by BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uq_flow_statuses_project_name (project_id, name),
    INDEX idx_flow_statuses_business_project (business_id, project_id, sort_order)
);

-- flow_tasks.status_id: NULL means "use the built-in string status".
SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
   WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'flow_tasks'
     AND COLUMN_NAME = 'status_id') = 0,
  'ALTER TABLE flow_tasks ADD COLUMN status_id BIGINT NULL AFTER status',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- Board queries filter by project and then by status, so the useful shape
-- mirrors the existing idx_flow_tasks_project_status rather than indexing
-- status_id alone.
SET @idx := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
   WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'flow_tasks'
     AND INDEX_NAME = 'idx_flow_tasks_project_status_id') = 0,
  'ALTER TABLE flow_tasks ADD INDEX idx_flow_tasks_project_status_id (project_id, status_id)',
  'SELECT 1'
);
PREPARE stmt2 FROM @idx; EXECUTE stmt2; DEALLOCATE PREPARE stmt2;

-- No foreign key, matching every other flow_* table. A task whose status row is
-- deleted must fall back to its string status rather than block the delete, and
-- the application clears status_id when a status is removed.
