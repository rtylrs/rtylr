-- name: ListSeriesByUserID :many
SELECT id, zontriq_user_id, name, description, niche, voice, style, frequency, state, video_count, config, publish_time, status, created_at, updated_at
FROM zontriq_series
WHERE zontriq_user_id = @zontriq_user_id
  AND status = 1
  AND (sqlc.narg('state')::TEXT IS NULL OR state = sqlc.narg('state')::TEXT)
ORDER BY created_at DESC
LIMIT @page_size OFFSET @page_offset;

-- name: CountSeriesByUserID :one
SELECT COUNT(*)::BIGINT AS total_count
FROM zontriq_series
WHERE zontriq_user_id = @zontriq_user_id
  AND status = 1
  AND (sqlc.narg('state')::TEXT IS NULL OR state = sqlc.narg('state')::TEXT);

-- name: GetSeriesByID :one
SELECT id, zontriq_user_id, name, description, niche, voice, style, frequency, state, video_count, config, publish_time, status, created_at, updated_at
FROM zontriq_series
WHERE id = @id
  AND status = 1;

-- name: CreateSeries :one
INSERT INTO zontriq_series (zontriq_user_id, name, description, niche, voice, style, frequency, config, publish_time)
VALUES (@zontriq_user_id, @name, @description, @niche, @voice, @style, @frequency, @config, @publish_time)
RETURNING id, zontriq_user_id, name, description, niche, voice, style, frequency, state, video_count, config, publish_time, status, created_at, updated_at;

-- name: UpdateSeries :one
UPDATE zontriq_series
SET name = @name,
    description = @description,
    niche = @niche,
    voice = @voice,
    style = @style,
    frequency = @frequency,
    config = @config,
    publish_time = @publish_time,
    state = @state,
    updated_at = NOW()
WHERE id = @id
  AND status = 1
RETURNING id, zontriq_user_id, name, description, niche, voice, style, frequency, state, video_count, config, publish_time, status, created_at, updated_at;

-- name: SoftDeleteSeries :exec
UPDATE zontriq_series
SET status = 0,
    updated_at = NOW()
WHERE id = @id
  AND status = 1;

-- name: IncrementVideoCount :exec
UPDATE zontriq_series
SET video_count = video_count + 1,
    updated_at = NOW()
WHERE id = @id
  AND status = 1;

-- name: GetSeriesContext :one
SELECT id, series_id, context, topics_covered, last_episode_number, status, created_at, updated_at
FROM zontriq_series_context
WHERE series_id = @series_id
  AND status = 1;

-- name: CreateSeriesContext :one
INSERT INTO zontriq_series_context (series_id, context)
VALUES (@series_id, @context)
RETURNING id, series_id, context, topics_covered, last_episode_number, status, created_at, updated_at;

-- name: UpdateSeriesContext :one
UPDATE zontriq_series_context
SET context = @context,
    updated_at = NOW()
WHERE series_id = @series_id
  AND status = 1
RETURNING id, series_id, context, topics_covered, last_episode_number, status, created_at, updated_at;

-- name: CountSeriesByUserIDForQuota :one
SELECT COUNT(*)::BIGINT AS total_count
FROM zontriq_series
WHERE zontriq_user_id = @zontriq_user_id
  AND status = 1;
