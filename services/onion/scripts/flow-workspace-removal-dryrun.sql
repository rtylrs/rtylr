-- Read-only. Run this before Phase 1's backfill (zz_flow_project_member_backfill.sql)
-- and again after, against the same database. Nothing here writes.
--
-- Every number should be 0 on the second run except total_projects and
-- cycles_without_project. If pairs_losing_access is still above 0 after the
-- backfill, do not proceed to the drops.

SELECT '--- 1. workspace members who would lose project access ---' AS section;
SELECT COUNT(*) AS member_pairs_losing_access
FROM flow_workspace_members wm
JOIN flow_projects p ON p.workspace_id = wm.workspace_id
LEFT JOIN flow_project_members pm
       ON pm.project_id = p.id AND pm.user_id = wm.user_id
WHERE pm.id IS NULL
  AND p.owner_id <> wm.user_id
  AND wm.user_id > 0;

-- The design doc's version of the query above also excluded
-- p.created_by <> wm.user_id. created_by grants nothing in projectRoleFor, only
-- owner_id does, so excluding it undercounts anyone who created a project they
-- do not own. This query drops that filter.

SELECT '--- 2. workspace OWNERS who would lose project access ---' AS section;
SELECT COUNT(*) AS owner_pairs_losing_access
FROM flow_workspaces w
JOIN flow_projects p ON p.workspace_id = w.id
LEFT JOIN flow_project_members pm
       ON pm.project_id = p.id AND pm.user_id = w.owner_id
WHERE pm.id IS NULL
  AND p.owner_id <> w.owner_id
  AND w.owner_id > 0;

-- A workspace owner holds no flow_workspace_members row (workspaceRoleFor
-- resolves them separately), so query 1 cannot see them. This category was
-- absent from the design doc's dry-run entirely.

SELECT '--- 3. combined blast radius ---' AS section;
SELECT COUNT(*) AS distinct_users_affected FROM (
  SELECT wm.user_id AS uid
  FROM flow_workspace_members wm
  JOIN flow_projects p ON p.workspace_id = wm.workspace_id
  LEFT JOIN flow_project_members pm
         ON pm.project_id = p.id AND pm.user_id = wm.user_id
  WHERE pm.id IS NULL AND p.owner_id <> wm.user_id AND wm.user_id > 0
  UNION
  SELECT w.owner_id AS uid
  FROM flow_workspaces w
  JOIN flow_projects p ON p.workspace_id = w.id
  LEFT JOIN flow_project_members pm
         ON pm.project_id = p.id AND pm.user_id = w.owner_id
  WHERE pm.id IS NULL AND p.owner_id <> w.owner_id AND w.owner_id > 0
) AS affected;

SELECT '--- 4. cycles with no project (Phase 4 step 11) ---' AS section;
SELECT COUNT(*) AS cycles_without_project
FROM flow_cycles
WHERE project_id IS NULL OR project_id = 0;

SELECT '--- 5. scale reference ---' AS section;
SELECT
  (SELECT COUNT(*) FROM flow_projects)           AS total_projects,
  (SELECT COUNT(*) FROM flow_workspaces)         AS total_workspaces,
  (SELECT COUNT(*) FROM flow_workspace_members)  AS total_workspace_members,
  (SELECT COUNT(*) FROM flow_project_members)    AS total_project_members;
