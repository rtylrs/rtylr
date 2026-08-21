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
-- Double counting is therefore not possible in payroll, which reads one store.
-- It is not prevented at write time either: the same hour may legitimately be
-- entered in both, and nothing reconciles them.
--
-- What IS wrong today is the reverse of double counting. Both per-task rollups
-- above read flow_time_logs alone, so a work-day line carrying a task_id
-- contributes nothing to that task's actual_hours or to insights. Teams using
-- the work-day screen, which is what this file exists for, see undercounts.
-- Tracked in tryrtylr/rtylr#903. Do not "fix" it by summing both stores without
-- reading that issue: the same hour can be in both.
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
-- live rows, and every existing entry predates work days.
ALTER TABLE hr_time_entries
    ADD COLUMN work_day_id BIGINT NULL,
    ADD COLUMN work_type_id BIGINT NULL,
    -- Null means inherit the day's mode; set means this line differed.
    ADD COLUMN work_mode VARCHAR(20) NULL,
    ADD COLUMN timezone VARCHAR(64) NULL,
    -- Client-generated before the line is sent, so an offline retry is idempotent.
    -- This is the difference between a field worker syncing twice and logging twice.
    ADD COLUMN client_key VARCHAR(64) NULL,
    ADD COLUMN logged_by BIGINT NULL;

ALTER TABLE hr_time_entries
    ADD INDEX idx_time_entries_work_day (owner_id, work_day_id);

-- Idempotency is only meaningful if the retry collides. Unique per tenant, and
-- NULL client_keys do not collide in MySQL, so legacy rows are unaffected.
ALTER TABLE hr_time_entries
    ADD UNIQUE KEY uk_time_entries_client_key (owner_id, client_key);
