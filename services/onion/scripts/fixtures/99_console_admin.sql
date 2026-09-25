-- Internal console admin account for console.voxire.com
-- Password omitted; fixture authentication is disabled.
INSERT INTO users (email, password_hash, full_name, onboarded, is_console_admin)
VALUES (
    'dev@voxire.com',
    'fixture-auth-disabled',
    'Voxire Dev',
    TRUE,
    TRUE
)
ON DUPLICATE KEY UPDATE
    is_console_admin = TRUE,
    password_hash    = VALUES(password_hash);

INSERT INTO console_stats (id)
VALUES (1)
ON DUPLICATE KEY UPDATE id = id;

UPDATE console_stats
SET total_users = (SELECT COUNT(*) FROM users),
    banned_users = (SELECT COUNT(*) FROM users WHERE is_banned = TRUE),
    total_businesses = (SELECT COUNT(*) FROM businesses),
    active_subs = (SELECT COUNT(*) FROM subscriptions WHERE status = 'active'),
    trial_subs = (SELECT COUNT(*) FROM subscriptions WHERE status = 'trial'),
    manual_grants = (SELECT COUNT(*) FROM manual_access_grants WHERE revoked_at IS NULL),
    manual_addon_grants = (SELECT COUNT(*) FROM manual_addon_grants WHERE revoked_at IS NULL),
    console_admins = (SELECT COUNT(*) FROM users WHERE is_console_admin = TRUE)
WHERE id = 1;
