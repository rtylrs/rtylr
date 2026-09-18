-- Phase 1 of the workspace removal claimed this backfill was already done. It
-- was not. Until it runs, project access works only because projectRoleFor
-- falls back through GetProjectOwnerWorkspace to the workspace roster
-- (service.go:288). Dropping flow_workspace_members without this silently
-- revokes access instead of failing loudly.
--
-- projectRoleFor resolves a caller's role on a project in this order:
--   1. flow_project_members row for (project, user)
--   2. flow_projects.owner_id = user            -> admin
--   3. flow_workspaces.owner_id = user          -> admin   (workspace owner)
--   4. flow_workspace_members row for (workspace, user)
--
-- Step 2 survives the drop because owner_id stays, so this file materialises
-- steps 3 and 4 as real project_members rows. Roles map 1:1 because both
-- tables share the same ENUM('admin','member','viewer') and a workspace role
-- already applies to every project in that workspace. A backfill preserves
-- current access; it is not the place to re-decide who should have what.
--
-- Guarded on INFORMATION_SCHEMA rather than written as a plain INSERT because
-- zz_ files re-run on every migrate pass. Once Phase 4 drops the workspace
-- tables this file has to become a no-op instead of a failure, and migrate.sh
-- records failures as applied behind a green check, so an unguarded version
-- would fail silently and forever.

SET @has_ws_members := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLES
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'flow_workspace_members'
);
SET @has_ws := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLES
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'flow_workspaces'
);
SET @has_project_ws_col := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'flow_projects'
    AND COLUMN_NAME = 'workspace_id'
);

-- Step 4: every workspace member gets their workspace role on every project in
-- that workspace. owner_id is excluded because step 2 already covers it.
SET @sql_members := IF(
  @has_ws_members > 0 AND @has_project_ws_col > 0,
  'INSERT IGNORE INTO flow_project_members (project_id, user_id, role, joined_at)
   SELECT p.id, wm.user_id, wm.role, wm.joined_at
   FROM flow_workspace_members wm
   JOIN flow_projects p ON p.workspace_id = wm.workspace_id
   LEFT JOIN flow_project_members pm
          ON pm.project_id = p.id AND pm.user_id = wm.user_id
   WHERE pm.id IS NULL
     AND p.owner_id <> wm.user_id
     AND wm.user_id > 0',
  'SELECT 1'
);
PREPARE stmt_members FROM @sql_members;
EXECUTE stmt_members;
DEALLOCATE PREPARE stmt_members;

-- Step 3: the workspace owner holds no flow_workspace_members row at all
-- (workspaceRoleFor resolves them separately, service.go:306), so a backfill
-- driven only off the members table leaves them with no project row and they
-- lose access to every project in their own workspace that they do not
-- personally own. They resolve to admin today, so that is the role recorded.
SET @sql_owners := IF(
  @has_ws > 0 AND @has_project_ws_col > 0,
  'INSERT IGNORE INTO flow_project_members (project_id, user_id, role, joined_at)
   SELECT p.id, w.owner_id, ''admin'', w.created_at
   FROM flow_workspaces w
   JOIN flow_projects p ON p.workspace_id = w.id
   LEFT JOIN flow_project_members pm
          ON pm.project_id = p.id AND pm.user_id = w.owner_id
   WHERE pm.id IS NULL
     AND p.owner_id <> w.owner_id
     AND w.owner_id > 0',
  'SELECT 1'
);
PREPARE stmt_owners FROM @sql_owners;
EXECUTE stmt_owners;
DEALLOCATE PREPARE stmt_owners;
