CREATE TABLE IF NOT EXISTS vox_user_permission_groups (
    user_id INTEGER NOT NULL,
    permission_group_id INTEGER NOT NULL,
    assigned_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    assigned_by INTEGER,
    PRIMARY KEY (user_id, permission_group_id)
);