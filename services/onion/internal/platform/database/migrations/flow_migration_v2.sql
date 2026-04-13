-- Flow schema migration v2
-- MySQL 5.7 compatible (no ADD COLUMN IF NOT EXISTS)
-- Uses information_schema checks via prepared statements.

-- ─────────────────────────────────────────────────────────────────────────────
-- flow_workspaces: add business_id
-- ─────────────────────────────────────────────────────────────────────────────
SET @db = DATABASE();

SET @s = IF(
  (SELECT COUNT(*) FROM information_schema.COLUMNS
   WHERE TABLE_SCHEMA = @db AND TABLE_NAME = 'flow_workspaces' AND COLUMN_NAME = 'business_id') = 0,
  'ALTER TABLE flow_workspaces ADD COLUMN business_id BIGINT NULL AFTER id',
  'SELECT 1'
);
PREPARE stmt FROM @s; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ─────────────────────────────────────────────────────────────────────────────
-- flow_workspaces: add is_personal
-- ─────────────────────────────────────────────────────────────────────────────
SET @s = IF(
  (SELECT COUNT(*) FROM information_schema.COLUMNS
   WHERE TABLE_SCHEMA = @db AND TABLE_NAME = 'flow_workspaces' AND COLUMN_NAME = 'is_personal') = 0,
  'ALTER TABLE flow_workspaces ADD COLUMN is_personal BOOLEAN NOT NULL DEFAULT FALSE AFTER is_default',
  'SELECT 1'
);
PREPARE stmt FROM @s; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- Seed is_personal from is_default for existing rows
UPDATE flow_workspaces SET is_personal = is_default WHERE is_personal = FALSE AND is_default = TRUE;

-- ─────────────────────────────────────────────────────────────────────────────
-- flow_workspace_members: change role ENUM + add invited_by
-- ─────────────────────────────────────────────────────────────────────────────

-- Remap any old 'admin' rows to 'member' before modifying enum
UPDATE flow_workspace_members SET role = 'member' WHERE role NOT IN ('manager','member','viewer');

ALTER TABLE flow_workspace_members
  MODIFY COLUMN role ENUM('manager','member','viewer') NOT NULL DEFAULT 'member';

SET @s = IF(
  (SELECT COUNT(*) FROM information_schema.COLUMNS
   WHERE TABLE_SCHEMA = @db AND TABLE_NAME = 'flow_workspace_members' AND COLUMN_NAME = 'invited_by') = 0,
  'ALTER TABLE flow_workspace_members ADD COLUMN invited_by BIGINT NULL AFTER role',
  'SELECT 1'
);
PREPARE stmt FROM @s; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ─────────────────────────────────────────────────────────────────────────────
-- flow_project_members: change role ENUM + add added_by
-- ─────────────────────────────────────────────────────────────────────────────

UPDATE flow_project_members SET role = 'member' WHERE role NOT IN ('manager','member','viewer');

ALTER TABLE flow_project_members
  MODIFY COLUMN role ENUM('manager','member','viewer') NOT NULL DEFAULT 'member';

SET @s = IF(
  (SELECT COUNT(*) FROM information_schema.COLUMNS
   WHERE TABLE_SCHEMA = @db AND TABLE_NAME = 'flow_project_members' AND COLUMN_NAME = 'added_by') = 0,
  'ALTER TABLE flow_project_members ADD COLUMN added_by BIGINT NULL AFTER role',
  'SELECT 1'
);
PREPARE stmt FROM @s; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ─────────────────────────────────────────────────────────────────────────────
-- flow_projects: make owner_id nullable (for ownership transfer)
-- ─────────────────────────────────────────────────────────────────────────────

ALTER TABLE flow_projects
  MODIFY COLUMN owner_id BIGINT NULL;
