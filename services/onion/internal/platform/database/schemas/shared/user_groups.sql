CREATE TABLE IF NOT EXISTS user_groups (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    color VARCHAR(7) DEFAULT '#6366f1',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_groups_owner (owner_id),
    UNIQUE KEY uk_group_owner_name (owner_id, name)
);

CREATE TABLE IF NOT EXISTS user_group_resource_permissions (
    group_id BIGINT NOT NULL,
    resource_key VARCHAR(64) NOT NULL,
    permission_level ENUM('viewer', 'editor') NOT NULL DEFAULT 'viewer',
    PRIMARY KEY (group_id, resource_key),
    INDEX idx_ugrp_group (group_id)
);

CREATE TABLE IF NOT EXISTS user_group_members (
    group_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    added_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    added_by BIGINT NULL,
    PRIMARY KEY (group_id, user_id),
    INDEX idx_ugm_user (user_id, owner_id),
    INDEX idx_ugm_group (group_id)
);

-- Per-group app permissions (group-level analogue of user_app_permissions).
CREATE TABLE IF NOT EXISTS user_group_app_permissions (
    group_id BIGINT NOT NULL,
    app_id VARCHAR(32) NOT NULL,
    permission_level ENUM('use','edit','own') NOT NULL DEFAULT 'use',
    PRIMARY KEY (group_id, app_id),
    INDEX idx_ugap_group (group_id)
);
