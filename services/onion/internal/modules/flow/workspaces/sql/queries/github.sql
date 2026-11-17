-- T-366: GitHub App installation per business and repositories per project.
-- Every statement is scoped by business_id. Project authorization is the
-- caller's job through RequireProjectRole.

-- name: GetGitHubInstallation :one
SELECT id, business_id, installation_id, account_login, account_type, installed_by,
       suspended_at, last_webhook_at, last_webhook_event, created_at
FROM flow_github_installations
WHERE business_id = ?;

-- GitHub webhooks name the installation, not the business. installation_id
-- is unique across businesses, so this resolves the one business that owns it.
-- name: GetGitHubInstallationByInstallationID :one
SELECT id, business_id, installation_id, account_login, account_type, installed_by,
       suspended_at, last_webhook_at, last_webhook_event, created_at
FROM flow_github_installations
WHERE installation_id = ?;

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

-- T-366 part B: webhook deliveries and installation health.

-- name: DeleteGitHubInstallationByID :exec
DELETE FROM flow_github_installations
WHERE business_id = ? AND installation_id = ?;

-- name: DeleteInstallationProjectRepositories :exec
DELETE FROM flow_project_repositories
WHERE business_id = ? AND installation_id = ?;

-- name: DeleteInstallationProjectRepository :exec
DELETE FROM flow_project_repositories
WHERE business_id = ? AND installation_id = ? AND repository_id = ?;

-- name: ListInstallationProjectRepositoryIDs :many
SELECT DISTINCT repository_id
FROM flow_project_repositories
WHERE business_id = ? AND installation_id = ?;

-- name: SetGitHubInstallationSuspendedAt :exec
UPDATE flow_github_installations
SET suspended_at = ?
WHERE business_id = ? AND installation_id = ?;

-- name: TouchGitHubInstallationWebhook :exec
UPDATE flow_github_installations
SET last_webhook_at = ?, last_webhook_event = ?
WHERE business_id = ? AND installation_id = ?;

-- Zero rows affected means the delivery id was already recorded.
-- name: RecordGitHubWebhookDelivery :execrows
INSERT IGNORE INTO flow_github_webhook_deliveries (delivery_id, business_id, event)
VALUES (?, ?, ?);

-- name: PruneGitHubWebhookDeliveries :exec
DELETE FROM flow_github_webhook_deliveries
WHERE received_at < ?
LIMIT 500;

-- T-366 part B: each member's linked GitHub account.

-- name: GetGitHubUserLink :one
SELECT business_id, user_id, github_user_id, github_login, github_avatar_url, created_at
FROM flow_github_user_links
WHERE business_id = ? AND user_id = ?;

-- name: CreateGitHubUserLink :exec
INSERT INTO flow_github_user_links
  (business_id, user_id, github_user_id, github_login, github_avatar_url)
VALUES (?, ?, ?, ?, ?);

-- name: DeleteGitHubUserLink :exec
DELETE FROM flow_github_user_links
WHERE business_id = ? AND user_id = ?;

-- A revoked authorization ends the GitHub user's grant to the app as a whole,
-- so the link goes in every business that holds it.
-- name: DeleteGitHubUserLinksByGitHubUser :execrows
DELETE FROM flow_github_user_links
WHERE github_user_id = ?;
