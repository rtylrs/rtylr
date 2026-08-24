CREATE TABLE IF NOT EXISTS flow_workspaces (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    icon VARCHAR(50) DEFAULT '🚀',
    color VARCHAR(20) DEFAULT '#4583C4',
    is_default BOOLEAN DEFAULT FALSE,
    is_personal BOOLEAN DEFAULT FALSE,
    created_by BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS flow_workspace_members (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    workspace_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    role ENUM('admin','member','viewer') DEFAULT 'member',
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY unique_member (workspace_id, user_id)
);

CREATE TABLE IF NOT EXISTS flow_projects (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    workspace_id BIGINT NOT NULL,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    icon VARCHAR(50) DEFAULT '📋',
    color VARCHAR(20) DEFAULT '#4583C4',
    status ENUM('active','archived','completed') DEFAULT 'active',
    start_date DATE,
    due_date DATE,
    created_by BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_flow_projects_business_created (business_id, created_at, id)
);

CREATE TABLE IF NOT EXISTS flow_project_members (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    project_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    role ENUM('admin','member','viewer') DEFAULT 'member',
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY unique_project_member (project_id, user_id)
);

CREATE TABLE IF NOT EXISTS flow_labels (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    workspace_id BIGINT NOT NULL,
    name VARCHAR(100) NOT NULL,
    color VARCHAR(20) NOT NULL DEFAULT '#6366f1',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS flow_tasks (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    project_id BIGINT NOT NULL,
    workspace_id BIGINT NOT NULL,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    parent_task_id BIGINT DEFAULT NULL,
    title VARCHAR(500) NOT NULL,
    description TEXT,
    status ENUM('backlog','todo','in_progress','in_review','done','cancelled') DEFAULT 'todo',
    -- F-A: the custom status this task carries, NULL meaning it uses the
    -- built-in string status above. Additive on purpose: every existing row
    -- is NULL and behaves exactly as before, and the string column stays
    -- populated either way so no existing query or index changes.
    status_id BIGINT NULL,
    priority ENUM('none','low','medium','high','urgent') DEFAULT 'none',
    assignee_id BIGINT,
    reporter_id BIGINT NOT NULL,
    due_date DATE,
    start_date DATE,
    completed_at TIMESTAMP NULL,
    estimated_hours DECIMAL(6,2),
    actual_hours DECIMAL(6,2),
    effort_minutes INT DEFAULT NULL,
    waiting_on VARCHAR(255) DEFAULT NULL,
    decision_note TEXT DEFAULT NULL,
    section_id BIGINT DEFAULT NULL,
    sort_order INT DEFAULT 0,
    archived BOOLEAN DEFAULT FALSE,
    labels JSON COMMENT 'Array of label IDs',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_flow_tasks_project_status (project_id, status),
    INDEX idx_flow_tasks_project_status_id (project_id, status_id),
    INDEX idx_flow_tasks_workspace_updated (workspace_id, archived, updated_at),
    INDEX idx_flow_tasks_business_status_completed (business_id, status, completed_at),
    INDEX idx_flow_tasks_business_assignee (business_id, assignee_id),
    INDEX idx_flow_tasks_project_parent (project_id, parent_task_id),
    FULLTEXT INDEX idx_flow_tasks_search (title, description)
);

CREATE TABLE IF NOT EXISTS flow_task_comments (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    task_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    content TEXT NOT NULL,
    is_edited BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS flow_task_activity (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    task_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    action VARCHAR(100) NOT NULL,
    old_value TEXT,
    new_value TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS flow_task_attachments (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    task_id BIGINT NOT NULL,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    name VARCHAR(255) NOT NULL,
    url TEXT NOT NULL,
    size BIGINT DEFAULT 0,
    uploaded_by BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS flow_time_logs (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    task_id BIGINT NOT NULL,
    -- When set, this row is the same hour as the named hr_time_entries row and
    -- the per-task rollups skip it. See flow/zz_flow_time_logs_time_entry_id.sql
    -- and tryrtylr/rtylr#903.
    time_entry_id BIGINT NULL,
    user_id BIGINT NOT NULL,
    duration_minutes INT NOT NULL DEFAULT 0,
    description TEXT,
    logged_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS flow_milestones (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL DEFAULT 0,
    business_id BIGINT NOT NULL DEFAULT 0,
    project_id BIGINT NOT NULL,
    workspace_id BIGINT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    target_date DATE,
    due_date DATE NULL,
    status ENUM('upcoming','active','completed','missed','planning','at_risk') DEFAULT 'upcoming',
    completed_at TIMESTAMP NULL,
    created_by BIGINT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS flow_cycles (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL DEFAULT 0,
    business_id BIGINT NOT NULL DEFAULT 0,
    workspace_id BIGINT NOT NULL,
    project_id BIGINT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    start_date DATE,
    end_date DATE,
    status ENUM('upcoming','active','completed') DEFAULT 'upcoming',
    created_by BIGINT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS flow_views (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL DEFAULT 0,
    business_id BIGINT NOT NULL DEFAULT 0,
    workspace_id BIGINT NOT NULL,
    name VARCHAR(255) NOT NULL,
    icon VARCHAR(50) DEFAULT 'eye',
    filters JSON,
    sort_by VARCHAR(100) DEFAULT '',
    sort_dir ENUM('asc','desc') DEFAULT 'asc',
    group_by VARCHAR(100) DEFAULT '',
    is_default BOOLEAN DEFAULT FALSE,
    visibility ENUM('private','shared') NOT NULL DEFAULT 'shared',
    created_by BIGINT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS flow_project_columns (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    project_id BIGINT NOT NULL,
    workspace_id BIGINT NOT NULL,
    name VARCHAR(255) NOT NULL,
    color VARCHAR(20) DEFAULT '#6366f1',
    sort_order INT DEFAULT 0,
    created_by BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_flow_columns_project (project_id)
);

CREATE TABLE IF NOT EXISTS flow_statuses (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    -- No DEFAULT 0, unlike the eleven sibling flow_* tables. A default of 0 is
    -- what made the task-dependency tenancy bug silent rather than an error:
    -- rows were written with business_id 0 and read with the caller's real id,
    -- so adding a dependency succeeded and it never appeared. Here a missing
    -- write fails loudly.
    business_id BIGINT NOT NULL,
    -- Status sets are per project. A project with no rows here falls back to
    -- the six built-in string statuses, which is what makes this additive:
    -- every existing project keeps working with zero rows written.
    project_id BIGINT NOT NULL,
    name VARCHAR(64) NOT NULL,
    -- The semantic bucket, for rollups that have to work across projects with
    -- different pipelines. Four buckets exactly as the proposal specifies.
    --
    -- The mapping for the six built-ins is not a fresh invention: the frontend
    -- already derives OPEN, CLOSED and ACTIVE sets in constantsFlow.ts, and
    -- these buckets are a refinement of those three. ACTIVE_STATUSES is
    -- in_progress and in_review only, so todo belongs to backlog, and then
    -- OPEN = backlog + active, CLOSED = done + cancelled. Every set the app
    -- computes today stays reconstructible.
    bucket ENUM('backlog','active','done','cancelled') NOT NULL,
    -- Which of the six enum values flow_tasks.status carries while this status
    -- is applied. The string column is not going away, so every query and
    -- index that filters on it keeps working; a custom status simply reports
    -- its nearest built-in. Built-in statuses map to themselves.
    legacy_status ENUM('backlog','todo','in_progress','in_review','done','cancelled') NOT NULL,
    color VARCHAR(20) NOT NULL DEFAULT '#6366f1',
    sort_order INT NOT NULL DEFAULT 0,
    -- The status a new task gets. At most one per project, enforced by the
    -- application: MySQL cannot express a partial unique index on it.
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    created_by BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uq_flow_statuses_project_name (project_id, name),
    INDEX idx_flow_statuses_business_project (business_id, project_id, sort_order)
);

-- Flow association + per-user tables (restored into the canonical schema after
-- the stale flow_migration_v*.sql files were removed; IF NOT EXISTS so applying
-- against the existing production tables is a no-op).
CREATE TABLE IF NOT EXISTS flow_businesses (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    name VARCHAR(255) NOT NULL,
    color VARCHAR(20) DEFAULT '#64C5B9',
    icon VARCHAR(50) DEFAULT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_flow_businesses_owner (owner_id)
);

CREATE TABLE IF NOT EXISTS flow_cycle_tasks (
    cycle_id BIGINT NOT NULL,
    task_id BIGINT NOT NULL,
    added_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (cycle_id, task_id)
);

CREATE TABLE IF NOT EXISTS flow_milestone_tasks (
    milestone_id BIGINT NOT NULL,
    task_id BIGINT NOT NULL,
    PRIMARY KEY (milestone_id, task_id)
);

CREATE TABLE IF NOT EXISTS flow_notifications (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    workspace_id BIGINT DEFAULT NULL,
    type ENUM('assigned','mention','comment','status_change','due_soon') NOT NULL,
    title VARCHAR(500) NOT NULL,
    description TEXT,
    task_id BIGINT DEFAULT NULL,
    task_title VARCHAR(500) DEFAULT NULL,
    actor_id BIGINT DEFAULT NULL,
    actor_initials VARCHAR(10) DEFAULT NULL,
    project VARCHAR(255) DEFAULT NULL,
    project_color VARCHAR(20) DEFAULT NULL,
    is_read BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_flow_notifications_user (user_id),
    INDEX idx_flow_notifications_user_read (user_id, is_read)
);

CREATE TABLE IF NOT EXISTS flow_user_settings (
    user_id BIGINT NOT NULL PRIMARY KEY,
    task_assigned BOOLEAN DEFAULT TRUE,
    task_comments BOOLEAN DEFAULT TRUE,
    mentions BOOLEAN DEFAULT TRUE,
    status_changes BOOLEAN DEFAULT FALSE,
    weekly_digest BOOLEAN DEFAULT TRUE,
    email_notifications BOOLEAN DEFAULT FALSE,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- Pinning a task to Today is per-user, not per-task: Today is a personal view,
-- so a task-level column would leak one member's planning into another's list.
CREATE TABLE IF NOT EXISTS flow_task_focus_pins (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    task_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uq_flow_task_focus_pins (task_id, user_id),
    INDEX idx_flow_task_focus_pins_user (user_id, business_id)
);

CREATE TABLE IF NOT EXISTS flow_task_dependencies (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    task_id BIGINT NOT NULL,
    depends_on_task_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uq_flow_task_dependencies (task_id, depends_on_task_id),
    INDEX idx_flow_task_dependencies_task (task_id, business_id),
    INDEX idx_flow_task_dependencies_depends_on (depends_on_task_id, business_id)
);
