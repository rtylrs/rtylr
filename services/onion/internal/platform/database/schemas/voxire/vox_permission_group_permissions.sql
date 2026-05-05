CREATE TABLE IF NOT EXISTS vox_permission_group_permissions (
    permission_group_id INTEGER NOT NULL,
    permission_id INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (permission_group_id, permission_id)
);
