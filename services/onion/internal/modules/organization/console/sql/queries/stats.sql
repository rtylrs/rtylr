-- name: GetStats :one
SELECT
    (SELECT COUNT(*) FROM users) AS total_users,
    (SELECT COUNT(*) FROM users WHERE is_banned = TRUE) AS banned_users,
    (SELECT COUNT(*) FROM businesses) AS total_businesses,
    (SELECT COUNT(*) FROM subscriptions WHERE status = 'active') AS active_subs,
    (SELECT COUNT(*) FROM subscriptions WHERE status = 'trial') AS trial_subs,
    (SELECT COUNT(*) FROM manual_access_grants WHERE revoked_at IS NULL) AS manual_grants,
    (SELECT COUNT(*) FROM manual_addon_grants WHERE revoked_at IS NULL) AS manual_addon_grants,
    (SELECT COUNT(*) FROM users WHERE is_console_admin = TRUE) AS console_admins;
