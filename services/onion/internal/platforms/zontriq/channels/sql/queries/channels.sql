-- name: ListChannelsByUserID :many
SELECT
    id,
    zontriq_user_id,
    platform,
    platform_channel_id,
    channel_name,
    channel_handle,
    channel_avatar_url,
    access_token,
    refresh_token,
    token_expires_at,
    connection_state,
    connected_at,
    status,
    created_at,
    updated_at
FROM zontriq_channels
WHERE zontriq_user_id = $1
  AND status = 1
  AND (sqlc.narg('platform')::TEXT IS NULL OR platform = sqlc.narg('platform')::TEXT)
ORDER BY connected_at DESC;

-- name: GetChannelByID :one
SELECT
    id,
    zontriq_user_id,
    platform,
    platform_channel_id,
    channel_name,
    channel_handle,
    channel_avatar_url,
    access_token,
    refresh_token,
    token_expires_at,
    connection_state,
    connected_at,
    status,
    created_at,
    updated_at
FROM zontriq_channels
WHERE id = $1
  AND status = 1;

-- name: CreateChannel :one
INSERT INTO zontriq_channels (
    zontriq_user_id,
    platform,
    platform_channel_id,
    channel_name,
    channel_handle,
    channel_avatar_url,
    access_token,
    refresh_token,
    token_expires_at,
    connection_state
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
) RETURNING
    id,
    zontriq_user_id,
    platform,
    platform_channel_id,
    channel_name,
    channel_handle,
    channel_avatar_url,
    access_token,
    refresh_token,
    token_expires_at,
    connection_state,
    connected_at,
    status,
    created_at,
    updated_at;

-- name: UpdateChannelTokens :one
UPDATE zontriq_channels
SET
    access_token = $2,
    refresh_token = $3,
    token_expires_at = $4,
    connection_state = $5,
    updated_at = NOW()
WHERE id = $1
  AND status = 1
RETURNING
    id,
    zontriq_user_id,
    platform,
    platform_channel_id,
    channel_name,
    channel_handle,
    channel_avatar_url,
    access_token,
    refresh_token,
    token_expires_at,
    connection_state,
    connected_at,
    status,
    created_at,
    updated_at;

-- name: UpdateChannelConnectionState :exec
UPDATE zontriq_channels
SET
    connection_state = $2,
    updated_at = NOW()
WHERE id = $1
  AND status = 1;

-- name: SoftDeleteChannel :exec
UPDATE zontriq_channels
SET
    status = 0,
    connection_state = 'disconnected',
    updated_at = NOW()
WHERE id = $1
  AND status = 1;

-- name: GetChannelByPlatformID :one
SELECT
    id,
    zontriq_user_id,
    platform,
    platform_channel_id,
    channel_name,
    channel_handle,
    channel_avatar_url,
    access_token,
    refresh_token,
    token_expires_at,
    connection_state,
    connected_at,
    status,
    created_at,
    updated_at
FROM zontriq_channels
WHERE zontriq_user_id = $1
  AND platform = $2
  AND platform_channel_id = $3
  AND status = 1;

-- name: CountChannelsByUserID :one
SELECT COUNT(*) FROM zontriq_channels
WHERE zontriq_user_id = $1
  AND status = 1;

-- name: GetZontriqUserIDByUserID :one
SELECT id
FROM zontriq_users
WHERE user_id = $1
  AND status = 1;
