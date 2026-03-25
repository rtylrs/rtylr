-- name: ListAutomationsByUserID :many
SELECT
    a.id,
    a.zontriq_user_id,
    a.series_id,
    s.name AS series_name,
    a.name,
    a.frequency,
    a.state,
    a.next_run,
    a.last_run,
    a.videos_generated,
    a.status,
    a.created_at,
    a.updated_at
FROM zontriq_automations a
JOIN zontriq_series s ON s.id = a.series_id
WHERE a.zontriq_user_id = @zontriq_user_id
  AND a.status = 1
  AND (sqlc.narg('state')::TEXT IS NULL OR a.state = sqlc.narg('state')::TEXT)
ORDER BY a.created_at DESC
LIMIT @page_size OFFSET @page_offset;

-- name: CountAutomationsByUserID :one
SELECT COUNT(*)::BIGINT AS total_count
FROM zontriq_automations
WHERE zontriq_user_id = @zontriq_user_id
  AND status = 1
  AND (sqlc.narg('state')::TEXT IS NULL OR state = sqlc.narg('state')::TEXT);

-- name: GetAutomationByID :one
SELECT
    a.id,
    a.zontriq_user_id,
    a.series_id,
    s.name AS series_name,
    a.name,
    a.frequency,
    a.state,
    a.next_run,
    a.last_run,
    a.videos_generated,
    a.status,
    a.created_at,
    a.updated_at
FROM zontriq_automations a
JOIN zontriq_series s ON s.id = a.series_id
WHERE a.id = @id
  AND a.status = 1;

-- name: CreateAutomation :one
INSERT INTO zontriq_automations (zontriq_user_id, series_id, name, frequency)
VALUES (@zontriq_user_id, @series_id, @name, @frequency)
RETURNING id, zontriq_user_id, series_id, name, frequency, state, next_run, last_run, videos_generated, status, created_at, updated_at;

-- name: UpdateAutomation :one
UPDATE zontriq_automations
SET name = @name,
    frequency = @frequency,
    state = @state,
    updated_at = NOW()
WHERE id = @id
  AND status = 1
RETURNING id, zontriq_user_id, series_id, name, frequency, state, next_run, last_run, videos_generated, status, created_at, updated_at;

-- name: UpdateAutomationNextRun :exec
UPDATE zontriq_automations
SET next_run = @next_run,
    last_run = @last_run,
    updated_at = NOW()
WHERE id = @id
  AND status = 1;

-- name: IncrementAutomationVideosGenerated :exec
UPDATE zontriq_automations
SET videos_generated = videos_generated + 1,
    updated_at = NOW()
WHERE id = @id
  AND status = 1;

-- name: SoftDeleteAutomation :exec
UPDATE zontriq_automations
SET status = 0,
    updated_at = NOW()
WHERE id = @id
  AND status = 1;

-- name: ListAutomationChannels :many
SELECT
    ac.id,
    ac.automation_id,
    ac.channel_id,
    c.channel_name,
    c.platform
FROM zontriq_automation_channels ac
JOIN zontriq_channels c ON c.id = ac.channel_id
WHERE ac.automation_id = @automation_id
  AND ac.status = 1;

-- name: CreateAutomationChannel :one
INSERT INTO zontriq_automation_channels (automation_id, channel_id)
VALUES (@automation_id, @channel_id)
RETURNING id, automation_id, channel_id, status, created_at;

-- name: DeleteAutomationChannels :exec
UPDATE zontriq_automation_channels
SET status = 0
WHERE automation_id = @automation_id
  AND status = 1;

-- name: GetDueAutomations :many
SELECT
    a.id,
    a.zontriq_user_id,
    a.series_id,
    s.name AS series_name,
    a.name,
    a.frequency,
    a.state,
    a.next_run,
    a.last_run,
    a.videos_generated,
    a.status,
    a.created_at,
    a.updated_at
FROM zontriq_automations a
JOIN zontriq_series s ON s.id = a.series_id
WHERE a.state = 'active'
  AND a.next_run <= NOW()
  AND a.status = 1
ORDER BY a.next_run ASC;

-- name: CountAutomationsByUserIDForQuota :one
SELECT COUNT(*)::BIGINT AS total_count
FROM zontriq_automations
WHERE zontriq_user_id = @zontriq_user_id
  AND state = 'active'
  AND status = 1;

-- name: GetZontriqUserIDByUserID :one
SELECT id
FROM zontriq_users
WHERE user_id = @user_id
  AND status = 1;
