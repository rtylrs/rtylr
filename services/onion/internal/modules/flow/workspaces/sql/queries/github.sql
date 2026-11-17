-- T-366: GitHub App installation per business and repositories per project.
-- Every statement is scoped by business_id. Project authorization is the
-- caller's job through RequireProjectRole.

-- name: GetGitHubInstallation :one
SELECT id, business_id, installation_id, account_login, account_type, installed_by, created_at
FROM flow_github_installations
WHERE business_id = ?;

-- name: CreateGitHubInstallation :exec
INSERT INTO flow_github_installations
  (business_id, installation_id, account_login, account_type, installed_by)
VALUES (?, ?, ?, ?, ?);

-- name: UpdateGitHubInstallationAccount :exec
UPDATE flow_github_installations
SET account_login = ?, account_type = ?
WHERE business_id = ? AND installation_id = ?;

-- name: DeleteGitHubInstallation :exec
DELETE FROM flow_github_installations
WHERE business_id = ?;

-- name: DeleteBusinessProjectRepositories :exec
DELETE FROM flow_project_repositories
WHERE business_id = ?;

-- name: ListProjectRepositories :many
SELECT repository_id, full_name, is_private, created_at
FROM flow_project_repositories
WHERE business_id = ? AND project_id = ?
ORDER BY full_name ASC, repository_id ASC;

-- name: DeleteProjectRepositories :exec
DELETE FROM flow_project_repositories
WHERE business_id = ? AND project_id = ?;

-- name: CreateProjectRepository :exec
INSERT INTO flow_project_repositories
  (business_id, project_id, installation_id, repository_id, full_name, is_private, added_by)
VALUES (?, ?, ?, ?, ?, ?, ?);
