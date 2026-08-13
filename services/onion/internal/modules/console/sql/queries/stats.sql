-- name: GetStats :one
SELECT
    total_users,
    banned_users,
    total_businesses,
    active_subs,
    trial_subs,
    manual_grants,
    manual_addon_grants,
    console_admins
FROM console_stats
WHERE id = 1;
