-- Flow daily time logging: the work day, its approval history, and the work-type
-- tree its lines are classified by.
--
-- The lines themselves are hr_time_entries. Flow does not keep a second store of
-- hours: that table already carries project, task, category, date, start and end,
-- billable, status and approver, and HR payroll reads it directly. What Flow adds
-- below are the columns that table lacks.
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
