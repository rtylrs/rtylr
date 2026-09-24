-- name: GetColumnsByProject :many
SELECT id, project_id, workspace_id, name, color, sort_order, created_by, created_at, updated_at
FROM flow_project_columns
WHERE project_id = ?
ORDER BY sort_order ASC
LIMIT ? OFFSET ?;

-- name: CreateColumn :execresult
INSERT INTO flow_project_columns (project_id, workspace_id, name, color, sort_order, created_by)
VALUES (?, ?, ?, ?, ?, ?);

-- Guarded on the project for the same reason DeleteColumn below is: a v2 column
-- carries workspace_id = 0, so the workspace pair alone matches nothing and the
-- rename silently affects no rows.
-- name: UpdateColumn :execresult
UPDATE flow_project_columns
SET name = ?, color = ?, sort_order = ?, updated_at = NOW()
WHERE flow_project_columns.id = ?
  AND (
    EXISTS (SELECT 1 FROM flow_project_members pm
            WHERE pm.project_id = flow_project_columns.project_id AND pm.user_id = ?)
    OR EXISTS (SELECT 1 FROM flow_projects p
               WHERE p.id = flow_project_columns.project_id
                 AND (p.owner_id = ? OR p.created_by = ?))
    OR EXISTS (SELECT 1 FROM flow_workspaces w WHERE w.id = flow_project_columns.workspace_id AND w.owner_id = ?)
    OR EXISTS (SELECT 1 FROM flow_workspace_members wm WHERE wm.workspace_id = flow_project_columns.workspace_id AND wm.user_id = ?)
  );

-- Runs first in the column delete path, so it carries the same guard as
-- DeleteColumn below. See the note there.
-- name: ClearTasksSectionID :exec
UPDATE flow_tasks
SET section_id = NULL
WHERE section_id = ?
  AND (
    EXISTS (SELECT 1 FROM flow_project_members pm
            WHERE pm.project_id = flow_tasks.project_id AND pm.user_id = ?)
    OR EXISTS (SELECT 1 FROM flow_projects p
               WHERE p.id = flow_tasks.project_id
                 AND (p.owner_id = ? OR p.created_by = ?))
    OR EXISTS (SELECT 1 FROM flow_workspaces w WHERE w.id = flow_tasks.workspace_id AND w.owner_id = ?)
    OR EXISTS (SELECT 1 FROM flow_workspace_members wm WHERE wm.workspace_id = flow_tasks.workspace_id AND wm.user_id = ?)
  );

-- Guarded on the project, not the workspace. Every v2 column carries
-- workspace_id = 0 and no workspace row has id 0, so the two workspace EXISTS
-- clauses below can never be satisfied: on their own they matched zero rows,
-- requireAffected turned that into ErrUnauthorized, and the endpoint returned
-- 403 to everyone including the business owner. The project clauses in front
-- are the shape MoveTaskToColumn already uses. The workspace pair is kept so
-- any surviving workspace-era column still deletes.
-- name: DeleteColumn :execresult
DELETE FROM flow_project_columns
WHERE flow_project_columns.id = ?
  AND (
    EXISTS (SELECT 1 FROM flow_project_members pm
            WHERE pm.project_id = flow_project_columns.project_id AND pm.user_id = ?)
    OR EXISTS (SELECT 1 FROM flow_projects p
               WHERE p.id = flow_project_columns.project_id
                 AND (p.owner_id = ? OR p.created_by = ?))
    OR EXISTS (SELECT 1 FROM flow_workspaces w WHERE w.id = flow_project_columns.workspace_id AND w.owner_id = ?)
    OR EXISTS (SELECT 1 FROM flow_workspace_members wm WHERE wm.workspace_id = flow_project_columns.workspace_id AND wm.user_id = ?)
  );

-- name: MoveTaskToColumn :execresult
UPDATE flow_tasks
SET section_id = sqlc.narg('section_id'), updated_at = NOW()
WHERE flow_tasks.id = ?
  AND (
    EXISTS (SELECT 1 FROM flow_project_members pm
            WHERE pm.project_id = flow_tasks.project_id AND pm.user_id = ?)
    OR EXISTS (SELECT 1 FROM flow_projects p
               WHERE p.id = flow_tasks.project_id
                 AND (p.owner_id = ? OR p.created_by = ?))
    OR EXISTS (SELECT 1 FROM flow_workspaces w WHERE w.id = flow_tasks.workspace_id AND w.owner_id = ?)
    OR EXISTS (SELECT 1 FROM flow_workspace_members wm WHERE wm.workspace_id = flow_tasks.workspace_id AND wm.user_id = ?)
  )
  AND (
    sqlc.arg('column_id') = 0
    OR EXISTS (
      SELECT 1 FROM flow_project_columns c
      WHERE c.id = sqlc.arg('column_id')
        AND (
          c.project_id = flow_tasks.project_id
          -- Legacy workspace-era rows have no reliable project boundary. Keep
          -- that compatibility path, but only for a real non-zero workspace;
          -- v2 rows use workspace_id=0 and must take the project branch above.
          OR (
            flow_tasks.workspace_id <> 0
            AND c.workspace_id = flow_tasks.workspace_id
          )
        )
    )
  );

-- name: SeedDefaultColumn :execresult
INSERT INTO flow_project_columns (project_id, workspace_id, name, color, sort_order, created_by)
VALUES (?, ?, ?, ?, ?, ?);
