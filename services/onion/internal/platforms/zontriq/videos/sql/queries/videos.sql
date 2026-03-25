-- name: ListVideosByUserID :many
SELECT v.*
FROM zontriq_videos v
WHERE v.zontriq_user_id = $1
  AND v.status = 1
  AND (sqlc.narg('series_id')::INTEGER IS NULL OR v.series_id = sqlc.narg('series_id')::INTEGER)
  AND (sqlc.narg('state')::TEXT IS NULL OR v.state = sqlc.narg('state')::TEXT)
ORDER BY v.created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountVideosByUserID :one
SELECT COUNT(*) AS total
FROM zontriq_videos v
WHERE v.zontriq_user_id = $1
  AND v.status = 1
  AND (sqlc.narg('series_id')::INTEGER IS NULL OR v.series_id = sqlc.narg('series_id')::INTEGER)
  AND (sqlc.narg('state')::TEXT IS NULL OR v.state = sqlc.narg('state')::TEXT);

-- name: GetVideoByID :one
SELECT v.*, COALESCE(s.name, '') AS series_name
FROM zontriq_videos v
LEFT JOIN zontriq_series s ON s.id = v.series_id
WHERE v.id = $1
  AND v.status = 1;

-- name: CreateVideo :one
INSERT INTO zontriq_videos (
    zontriq_user_id,
    series_id,
    title,
    script,
    state,
    config
) VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: UpdateVideoState :exec
UPDATE zontriq_videos
SET state = $2, updated_at = NOW()
WHERE id = $1
  AND status = 1;

-- name: SoftDeleteVideo :exec
UPDATE zontriq_videos
SET status = 0, updated_at = NOW()
WHERE id = $1
  AND status = 1;

-- name: ListVideoTargetsByVideoID :many
SELECT
    vt.id,
    vt.video_id,
    vt.channel_id,
    vt.state,
    vt.published_at,
    vt.platform_post_id,
    vt.error_message,
    vt.status,
    vt.created_at,
    c.channel_name,
    c.platform
FROM zontriq_video_targets vt
JOIN zontriq_channels c ON c.id = vt.channel_id
WHERE vt.video_id = $1
  AND vt.status = 1;

-- name: CreateVideoTarget :one
INSERT INTO zontriq_video_targets (
    video_id,
    channel_id
) VALUES ($1, $2)
RETURNING *;

-- name: UpdateVideoTargetState :exec
UPDATE zontriq_video_targets
SET state = $2,
    platform_post_id = $3,
    published_at = $4
WHERE id = $1
  AND status = 1;

-- name: GetActiveJobByVideoID :one
SELECT
    j.id,
    j.video_id,
    j.type,
    j.state,
    j.progress,
    j.current_step,
    j.error_message,
    j.gpu_id,
    j.started_at,
    j.completed_at,
    j.status,
    j.created_at,
    j.updated_at
FROM zontriq_jobs j
WHERE j.video_id = $1
  AND j.state IN ('queued', 'running')
  AND j.status = 1
ORDER BY j.created_at DESC
LIMIT 1;

-- name: CreateJob :one
INSERT INTO zontriq_jobs (
    video_id,
    type,
    state
) VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdateJobProgress :exec
UPDATE zontriq_jobs
SET state = $2,
    progress = $3,
    current_step = $4,
    updated_at = NOW()
WHERE id = $1
  AND status = 1;

-- name: UpdateJobCompleted :exec
UPDATE zontriq_jobs
SET state = $2,
    completed_at = NOW(),
    error_message = $3,
    updated_at = NOW()
WHERE id = $1
  AND status = 1;

-- name: ListPublishLogByVideoID :many
SELECT
    pl.id,
    pl.video_id,
    pl.channel_id,
    pl.platform,
    pl.platform_post_id,
    pl.platform_post_url,
    pl.result,
    pl.error_message,
    pl.published_at,
    pl.status,
    pl.created_at
FROM zontriq_publish_log pl
WHERE pl.video_id = $1
  AND pl.status = 1
ORDER BY pl.published_at DESC;

-- name: CreatePublishLogEntry :one
INSERT INTO zontriq_publish_log (
    video_id,
    channel_id,
    platform,
    platform_post_id,
    platform_post_url,
    result,
    error_message
) VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetZontriqUserIDByUserID :one
SELECT id
FROM zontriq_users
WHERE user_id = $1
  AND status = 1;
