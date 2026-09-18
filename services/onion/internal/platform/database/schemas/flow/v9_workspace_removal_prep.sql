-- Workspace-removal groundwork: projects become first-class business citizens.
-- visibility powers private-by-default (creator + members) with an optional
-- public profile-listing mode; is_personal marks the auto-provisioned My Tasks
-- bucket so code stops guessing by name and emoji; the archive timestamps turn
-- project deletion into a recycle bin stamped at archive time from the
-- business's retention setting.
--
-- Single-apply like v6-v8: migrate.sh tracks applied files. The backfill gives
-- every pre-existing row private visibility and marks My Tasks buckets exactly,
-- via the personal-workspace join instead of name heuristics.

ALTER TABLE flow_projects
  ADD COLUMN visibility ENUM('private','public') NOT NULL DEFAULT 'private';

ALTER TABLE flow_projects
  ADD COLUMN is_personal BOOLEAN NOT NULL DEFAULT FALSE;

ALTER TABLE flow_projects
  ADD COLUMN archived_at TIMESTAMP NULL DEFAULT NULL;

ALTER TABLE flow_projects
  ADD COLUMN purge_at TIMESTAMP NULL DEFAULT NULL;

ALTER TABLE flow_projects
  ADD INDEX idx_flow_projects_business_visibility (business_id, visibility, status);

ALTER TABLE flow_projects
  ADD INDEX idx_flow_projects_purge (purge_at);

UPDATE flow_projects p
LEFT JOIN flow_workspaces w ON p.workspace_id = w.id AND w.is_personal = TRUE
SET p.visibility = 'private', p.is_personal = (w.id IS NOT NULL)
WHERE p.is_personal = FALSE;


ALTER TABLE flow_labels
  ADD COLUMN business_id BIGINT NOT NULL DEFAULT 0;

UPDATE flow_labels l
JOIN flow_workspaces w ON l.workspace_id = w.id
SET l.business_id = w.business_id
WHERE l.business_id = 0;
