-- One-off repair for two bugs where flow_projects/flow_workspaces rows never
-- got a real business_id: EnsureMyTasksProject wrote the caller's user id
-- into flow_projects.business_id instead of their tenant id, and
-- CreatePersonalWorkspace never set flow_workspaces.business_id at all,
-- leaving it at the schema default of 0. Both are fixed going forward in
-- application code; this repairs rows written before that fix.
--
-- Run the SELECT sections first and review the affected counts/rows. Only
-- run the UPDATE sections after that review - do not run this file
-- unattended against production.

-- ── Dry run: personal-workspace projects with the wrong business_id ────────
SELECT p.id AS project_id, p.owner_id, p.business_id AS current_business_id,
       u.business_id AS correct_business_id
FROM flow_projects p
JOIN flow_workspaces w ON p.workspace_id = w.id AND w.is_personal = 1
JOIN users u ON u.id = p.owner_id
WHERE u.business_id IS NOT NULL AND p.business_id != u.business_id;

-- ── Dry run: personal workspaces stuck at business_id = 0 ──────────────────
SELECT w.id AS workspace_id, w.owner_id, w.business_id AS current_business_id,
       u.business_id AS correct_business_id
FROM flow_workspaces w
JOIN users u ON u.id = w.owner_id
WHERE w.is_personal = 1 AND u.business_id IS NOT NULL AND w.business_id != u.business_id;

-- ── Repair: personal-workspace projects (run only after reviewing above) ──
-- UPDATE flow_projects p
-- JOIN flow_workspaces w ON p.workspace_id = w.id AND w.is_personal = 1
-- JOIN users u ON u.id = p.owner_id
-- SET p.business_id = u.business_id
-- WHERE u.business_id IS NOT NULL AND p.business_id != u.business_id;

-- ── Repair: personal workspaces (run only after reviewing above) ──────────
-- UPDATE flow_workspaces w
-- JOIN users u ON u.id = w.owner_id
-- SET w.business_id = u.business_id
-- WHERE w.is_personal = 1 AND u.business_id IS NOT NULL AND w.business_id != u.business_id;
