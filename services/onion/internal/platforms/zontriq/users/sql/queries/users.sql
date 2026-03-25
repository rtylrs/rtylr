-- name: GetZontriqUserByUserID :one
SELECT
    zu.*,
    u.email AS user_email,
    u.full_name AS user_name,
    zp.id AS plan_pk,
    zp.name AS plan_name,
    zp.display_name AS plan_display_name,
    zp.price_cents AS plan_price_cents,
    zp.max_videos_per_month AS plan_max_videos_per_month,
    zp.max_channels AS plan_max_channels,
    zp.max_series AS plan_max_series,
    zp.max_automations AS plan_max_automations
FROM zontriq_users zu
JOIN users u ON u.id = zu.user_id
JOIN zontriq_plans zp ON zp.id = zu.plan_id
WHERE zu.user_id = $1
  AND zu.status = 1;

-- name: GetUserPlanByUserID :one
SELECT zp.*
FROM zontriq_plans zp
JOIN zontriq_users zu ON zu.plan_id = zp.id
WHERE zu.user_id = $1
  AND zu.status = 1
  AND zp.status = 1;

-- name: CreateZontriqUser :one
INSERT INTO zontriq_users (user_id, plan_id)
VALUES ($1, $2)
RETURNING *;

-- name: UpdateZontriqUserPlan :exec
UPDATE zontriq_users
SET plan_id = $2, updated_at = NOW()
WHERE id = $1
  AND status = 1;

-- name: SoftDeleteZontriqUser :exec
UPDATE zontriq_users
SET status = 0, updated_at = NOW()
WHERE id = $1
  AND status = 1;

-- name: ListPlans :many
SELECT *
FROM zontriq_plans
WHERE status = 1
ORDER BY price_cents ASC;

-- name: GetPlanByID :one
SELECT *
FROM zontriq_plans
WHERE id = $1
  AND status = 1;

-- name: GetCurrentUsage :one
SELECT *
FROM zontriq_usage
WHERE zontriq_user_id = $1
  AND period_end > NOW()
  AND status = 1;

-- name: IncrementVideosGenerated :exec
UPDATE zontriq_usage
SET videos_generated = videos_generated + 1, updated_at = NOW()
WHERE id = $1
  AND status = 1;

-- name: IncrementVideosPublished :exec
UPDATE zontriq_usage
SET videos_published = videos_published + 1, updated_at = NOW()
WHERE id = $1
  AND status = 1;

-- name: CreateUsagePeriod :one
INSERT INTO zontriq_usage (zontriq_user_id, period_start, period_end)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetPreferences :one
SELECT *
FROM zontriq_preferences
WHERE zontriq_user_id = $1
  AND status = 1;

-- name: CreatePreferences :one
INSERT INTO zontriq_preferences (
    zontriq_user_id,
    default_voice,
    default_niche,
    notify_video_ready,
    notify_publish_success,
    notify_publish_failure,
    notify_quota_warning,
    notify_weekly_report
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: UpdatePreferences :one
UPDATE zontriq_preferences
SET default_voice = $2,
    default_niche = $3,
    notify_video_ready = $4,
    notify_publish_success = $5,
    notify_publish_failure = $6,
    notify_quota_warning = $7,
    notify_weekly_report = $8,
    updated_at = NOW()
WHERE zontriq_user_id = $1
  AND status = 1
RETURNING *;

-- name: CompleteOnboarding :exec
UPDATE zontriq_users
SET onboarded = TRUE, updated_at = NOW()
WHERE user_id = $1
  AND status = 1;

-- name: CreateWaitlistEntry :exec
INSERT INTO waitlist (email, source)
VALUES ($1, $2)
ON CONFLICT (email) DO NOTHING;
