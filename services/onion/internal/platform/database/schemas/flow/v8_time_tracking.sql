-- Flow daily time logging: the work day, its approval history, and the work-type
-- tree its lines are classified by.
--
-- The lines themselves are hr_time_entries: that table already carries project,
-- task, category, date, start and end, billable, status and approver, and HR
-- payroll reads it directly. What Flow adds below are the columns it lacks.
--
-- Flow does keep a second store of hours, and this comment used to deny it.
-- flow_time_logs (flow/schema.sql) holds per-task durations written by
-- flow.task.v1 CreateTimeLog, which the task detail screen's time panel calls.
-- The two stores mean different things and are read by different code:
--
--   hr_time_entries   the day's work, one row per line, source='flow' for ours.
--                     AUTHORITATIVE FOR PAYROLL. Carries task_id, so its rows
--                     are attributable per task.
--   flow_time_logs    ad-hoc "N minutes on this task" entries. Read only by
--                     Flow, for flow_tasks.actual_hours (RecalcActualHours) and
--                     the insights time-by-member panel. Payroll never sees it.
--
-- Double counting is not possible in payroll, which reads one store.
--
-- The per-task rollups read BOTH, as of tryrtylr/rtylr#903. They used to read
-- flow_time_logs alone, so a work-day line carrying a task_id contributed
-- nothing to that task's actual_hours or to insights, and teams using the
-- work-day screen this file exists for saw zeros.
--
-- Summing both needs an overlap key, because the same hour may legitimately be
-- entered in each. flow_time_logs.time_entry_id is it: when set, that row is the
-- same hour as the named hr_time_entries row and the rollups skip it, counting
-- the hr_time_entries side because payroll reads it. Nothing writes the column
-- yet, so today the rollups are a plain sum; linking the two at write time is
-- still open on #903.
--
-- History: the denial above was cited by tryrtylr/rtylr#869 and by
-- docs/cross-app/flow-deep-dive.md as proof that Flow's hours are a single
-- store. See tryrtylr/rtylr#890.
--
-- business_id is NOT NULL with no default on every new table here, deliberately.
-- The older tables default it to 0, which turns a forgotten value into a silently
-- shared tenant rather than an error. Fixing those is a separate change and needs
-- a data audit first; this file simply stops adding to the problem.

CREATE TABLE IF NOT EXISTS flow_work_days (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL,
    employee_id BIGINT NOT NULL,
    -- The calendar day in the person's own zone, which is why the zone is stored
    -- beside it: a travelling worker's day boundary is theirs, not the workspace's.
    date DATE NOT NULL,
    timezone VARCHAR(64) NOT NULL,
    -- onsite | remote | hybrid. The day's default; a line may override it.
    work_mode VARCHAR(20) NOT NULL DEFAULT 'onsite',
    -- Local HH:MM. Null for a day that records only its lines.
    start_time VARCHAR(5) NULL,
    end_time VARCHAR(5) NULL,
    -- Unpaid, so paid hours are not overstated.
    break_minutes INT NOT NULL DEFAULT 0,
    -- draft | submitted | approved | rejected | reopened
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    -- Differs from employee_id when a manager records the day for someone else.
    logged_by BIGINT NULL,
    approved_by BIGINT NULL,
    approved_at TIMESTAMP NULL,
    reopened_by BIGINT NULL,
    reopened_at TIMESTAMP NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    -- One day per person per date. The upsert depends on this.
    UNIQUE KEY uk_work_day (owner_id, business_id, employee_id, date),
    INDEX idx_work_days_tenant_employee_date (owner_id, business_id, employee_id, date),
    INDEX idx_work_days_tenant_status (owner_id, business_id, status, date)
);

-- Every approve, reject and reopen, so an approved day can be explained months
-- later. Mirrors hr_timesheet_approval_actions rather than reusing it, whose
-- timesheet_id means a period rather than a day.
CREATE TABLE IF NOT EXISTS flow_work_day_actions (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL,
    work_day_id BIGINT NOT NULL,
    actor_id BIGINT NOT NULL,
    -- submitted | approved | rejected | reopened
    action VARCHAR(20) NOT NULL,
    comment TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_work_day_actions_day (owner_id, business_id, work_day_id, created_at)
);

-- Exactly two levels: a top-level type with parent_id NULL, and its children.
-- hr_category_id maps a node onto HR's payroll category so one entry serves both
-- project reporting and payroll without a second classification on the line.
CREATE TABLE IF NOT EXISTS flow_work_types (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL,
    workspace_id BIGINT NOT NULL,
    parent_id BIGINT NULL,
    name VARCHAR(255) NOT NULL,
    -- Shown to employees as a read-only "Task Description" once they pick this
    -- type in Log Hours - added after this file was first applied everywhere,
    -- see zz_flow_work_types_description.sql for the idempotent live-DB patch.
    description TEXT,
    hr_category_id BIGINT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    -- Archived rather than deleted: historic entries keep pointing at it.
    archived BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_work_types_tenant_workspace (owner_id, business_id, workspace_id, archived, sort_order),
    INDEX idx_work_types_parent (owner_id, business_id, parent_id)
);

-- The line columns hr_time_entries lacks. All nullable or defaulted: the table has
-- live rows, and every existing entry predates work days. MySQL 8.4 has no
-- portable ADD COLUMN IF NOT EXISTS, so guard each alteration explicitly.
SET @sql := IF((SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hr_time_entries' AND COLUMN_NAME = 'work_day_id') = 0, 'ALTER TABLE hr_time_entries ADD COLUMN work_day_id BIGINT NULL', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql := IF((SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hr_time_entries' AND COLUMN_NAME = 'work_type_id') = 0, 'ALTER TABLE hr_time_entries ADD COLUMN work_type_id BIGINT NULL', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
-- Null means inherit the day's mode; set means this line differed.
SET @sql := IF((SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hr_time_entries' AND COLUMN_NAME = 'work_mode') = 0, 'ALTER TABLE hr_time_entries ADD COLUMN work_mode VARCHAR(20) NULL', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql := IF((SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hr_time_entries' AND COLUMN_NAME = 'timezone') = 0, 'ALTER TABLE hr_time_entries ADD COLUMN timezone VARCHAR(64) NULL', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
-- Client-generated before the line is sent, so an offline retry is idempotent.
-- This is the difference between a field worker syncing twice and logging twice.
SET @sql := IF((SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hr_time_entries' AND COLUMN_NAME = 'client_key') = 0, 'ALTER TABLE hr_time_entries ADD COLUMN client_key VARCHAR(64) NULL', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql := IF((SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hr_time_entries' AND COLUMN_NAME = 'logged_by') = 0, 'ALTER TABLE hr_time_entries ADD COLUMN logged_by BIGINT NULL', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql := IF((SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hr_time_entries' AND INDEX_NAME = 'idx_time_entries_work_day') = 0, 'ALTER TABLE hr_time_entries ADD INDEX idx_time_entries_work_day (owner_id, work_day_id)', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- Idempotency is only meaningful if the retry collides. Unique per tenant, and
-- NULL client_keys do not collide in MySQL, so legacy rows are unaffected.
SET @sql := IF((SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hr_time_entries' AND INDEX_NAME = 'uk_time_entries_client_key') = 0, 'ALTER TABLE hr_time_entries ADD UNIQUE KEY uk_time_entries_client_key (owner_id, client_key)', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
