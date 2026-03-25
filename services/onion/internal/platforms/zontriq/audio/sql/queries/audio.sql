-- name: CreateAudioClip :one
INSERT INTO zontriq_audio (
    zontriq_user_id,
    title,
    script,
    voice_id,
    voice_name,
    provider,
    state
) VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: ListAudioByUserID :many
SELECT a.*
FROM zontriq_audio a
WHERE a.zontriq_user_id = $1
  AND a.status = 1
ORDER BY a.created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountAudioByUserID :one
SELECT COUNT(*) AS total
FROM zontriq_audio a
WHERE a.zontriq_user_id = $1
  AND a.status = 1;

-- name: GetAudioByID :one
SELECT a.*
FROM zontriq_audio a
WHERE a.id = $1
  AND a.status = 1;

-- name: UpdateAudioState :exec
UPDATE zontriq_audio
SET state = $2,
    audio_url = $3,
    duration_seconds = $4,
    error_message = $5,
    updated_at = NOW()
WHERE id = $1
  AND status = 1;

-- name: SetExternalJobID :exec
UPDATE zontriq_audio
SET external_job_id = $2,
    state = 'generating',
    updated_at = NOW()
WHERE id = $1
  AND status = 1;

-- name: GetGeneratingAudio :many
SELECT a.*
FROM zontriq_audio a
WHERE a.state = 'generating'
  AND a.external_job_id IS NOT NULL
  AND a.status = 1
ORDER BY a.created_at ASC
LIMIT 10;

-- name: IncrementRetryCount :exec
UPDATE zontriq_audio
SET retry_count = retry_count + 1,
    state = 'queued',
    error_message = $2,
    updated_at = NOW()
WHERE id = $1
  AND status = 1;

-- name: SoftDeleteAudio :exec
UPDATE zontriq_audio
SET status = 0, updated_at = NOW()
WHERE id = $1
  AND status = 1;

-- name: GetQueuedAudio :many
SELECT a.*
FROM zontriq_audio a
WHERE a.state = 'queued'
  AND a.status = 1
ORDER BY a.created_at ASC
LIMIT 10;

-- name: GetZontriqUserIDByUserID :one
SELECT id
FROM zontriq_users
WHERE user_id = $1
  AND status = 1;
