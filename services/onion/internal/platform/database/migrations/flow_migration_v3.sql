-- Flow schema migration v3
-- Adds: cycles, milestones, views, notifications, user_settings

CREATE TABLE IF NOT EXISTS flow_cycles (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    workspace_id BIGINT NOT NULL,
    project_id BIGINT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    status ENUM('upcoming','active','completed') DEFAULT 'upcoming',
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    created_by BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_cycles_workspace (workspace_id)
);

CREATE TABLE IF NOT EXISTS flow_cycle_tasks (
    cycle_id BIGINT NOT NULL,
    task_id BIGINT NOT NULL,
    added_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (cycle_id, task_id)
);

CREATE TABLE IF NOT EXISTS flow_milestones (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    project_id BIGINT NOT NULL,
    workspace_id BIGINT NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    status ENUM('planning','active','at_risk','completed') DEFAULT 'planning',
    due_date DATE NOT NULL,
    completed_at TIMESTAMP NULL,
    created_by BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_milestones_project (project_id),
    INDEX idx_milestones_workspace (workspace_id)
);

CREATE TABLE IF NOT EXISTS flow_milestone_tasks (
    milestone_id BIGINT NOT NULL,
    task_id BIGINT NOT NULL,
    PRIMARY KEY (milestone_id, task_id)
);

CREATE TABLE IF NOT EXISTS flow_views (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    workspace_id BIGINT NOT NULL,
    name VARCHAR(255) NOT NULL,
    icon VARCHAR(50) DEFAULT 'eye',
    filters JSON,
    group_by VARCHAR(50),
    sort_by VARCHAR(50),
    sort_dir ENUM('asc','desc') DEFAULT 'asc',
    created_by BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_views_workspace (workspace_id)
);

CREATE TABLE IF NOT EXISTS flow_notifications (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    workspace_id BIGINT NULL,
    type ENUM('assigned','mention','comment','status_change','due_soon') NOT NULL,
    title VARCHAR(500) NOT NULL,
    description TEXT,
    task_id BIGINT NULL,
    task_title VARCHAR(500),
    actor_id BIGINT NULL,
    actor_initials VARCHAR(10),
    project VARCHAR(255),
    project_color VARCHAR(20),
    is_read BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_notifications_user (user_id),
    INDEX idx_notifications_user_read (user_id, is_read)
);

CREATE TABLE IF NOT EXISTS flow_user_settings (
    user_id BIGINT NOT NULL PRIMARY KEY,
    task_assigned BOOLEAN DEFAULT TRUE,
    task_comments BOOLEAN DEFAULT TRUE,
    mentions BOOLEAN DEFAULT TRUE,
    status_changes BOOLEAN DEFAULT FALSE,
    weekly_digest BOOLEAN DEFAULT TRUE,
    email_notifications BOOLEAN DEFAULT FALSE,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
