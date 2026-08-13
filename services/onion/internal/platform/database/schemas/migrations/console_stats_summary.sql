-- This migration lives outside SQLC schema inputs because it installs triggers.
-- Backfill the singleton once for existing installations. Subsequent writes are
-- tracked by the triggers below, so console stats are O(1) to read.
CREATE TABLE IF NOT EXISTS console_stats (
    id TINYINT NOT NULL PRIMARY KEY,
    total_users BIGINT NOT NULL DEFAULT 0,
    banned_users BIGINT NOT NULL DEFAULT 0,
    total_businesses BIGINT NOT NULL DEFAULT 0,
    active_subs BIGINT NOT NULL DEFAULT 0,
    trial_subs BIGINT NOT NULL DEFAULT 0,
    manual_grants BIGINT NOT NULL DEFAULT 0,
    manual_addon_grants BIGINT NOT NULL DEFAULT 0,
    console_admins BIGINT NOT NULL DEFAULT 0,
    CHECK (id = 1)
);

INSERT INTO console_stats (
    id, total_users, banned_users, total_businesses, active_subs, trial_subs,
    manual_grants, manual_addon_grants, console_admins
)
SELECT
    1,
    (SELECT COUNT(*) FROM users),
    (SELECT COUNT(*) FROM users WHERE is_banned = TRUE),
    (SELECT COUNT(*) FROM businesses),
    (SELECT COUNT(*) FROM subscriptions WHERE status = 'active'),
    (SELECT COUNT(*) FROM subscriptions WHERE status = 'trial'),
    (SELECT COUNT(*) FROM manual_access_grants WHERE revoked_at IS NULL),
    (SELECT COUNT(*) FROM manual_addon_grants WHERE revoked_at IS NULL),
    (SELECT COUNT(*) FROM users WHERE is_console_admin = TRUE)
ON DUPLICATE KEY UPDATE id = id;

DROP TRIGGER IF EXISTS console_stats_users_ai;
CREATE TRIGGER console_stats_users_ai AFTER INSERT ON users
FOR EACH ROW UPDATE console_stats
SET total_users = total_users + 1,
    banned_users = banned_users + (NEW.is_banned = TRUE),
    console_admins = console_admins + (NEW.is_console_admin = TRUE)
WHERE id = 1;

DROP TRIGGER IF EXISTS console_stats_users_au;
CREATE TRIGGER console_stats_users_au AFTER UPDATE ON users
FOR EACH ROW UPDATE console_stats
SET banned_users = banned_users + (NEW.is_banned = TRUE) - (OLD.is_banned = TRUE),
    console_admins = console_admins + (NEW.is_console_admin = TRUE) - (OLD.is_console_admin = TRUE)
WHERE id = 1;

DROP TRIGGER IF EXISTS console_stats_users_ad;
CREATE TRIGGER console_stats_users_ad AFTER DELETE ON users
FOR EACH ROW UPDATE console_stats
SET total_users = total_users - 1,
    banned_users = banned_users - (OLD.is_banned = TRUE),
    console_admins = console_admins - (OLD.is_console_admin = TRUE)
WHERE id = 1;

DROP TRIGGER IF EXISTS console_stats_businesses_ai;
CREATE TRIGGER console_stats_businesses_ai AFTER INSERT ON businesses
FOR EACH ROW UPDATE console_stats SET total_businesses = total_businesses + 1 WHERE id = 1;

DROP TRIGGER IF EXISTS console_stats_businesses_ad;
CREATE TRIGGER console_stats_businesses_ad AFTER DELETE ON businesses
FOR EACH ROW UPDATE console_stats SET total_businesses = total_businesses - 1 WHERE id = 1;

DROP TRIGGER IF EXISTS console_stats_subscriptions_ai;
CREATE TRIGGER console_stats_subscriptions_ai AFTER INSERT ON subscriptions
FOR EACH ROW UPDATE console_stats
SET active_subs = active_subs + (NEW.status = 'active'),
    trial_subs = trial_subs + (NEW.status = 'trial')
WHERE id = 1;

DROP TRIGGER IF EXISTS console_stats_subscriptions_au;
CREATE TRIGGER console_stats_subscriptions_au AFTER UPDATE ON subscriptions
FOR EACH ROW UPDATE console_stats
SET active_subs = active_subs + (NEW.status = 'active') - (OLD.status = 'active'),
    trial_subs = trial_subs + (NEW.status = 'trial') - (OLD.status = 'trial')
WHERE id = 1;

DROP TRIGGER IF EXISTS console_stats_subscriptions_ad;
CREATE TRIGGER console_stats_subscriptions_ad AFTER DELETE ON subscriptions
FOR EACH ROW UPDATE console_stats
SET active_subs = active_subs - (OLD.status = 'active'),
    trial_subs = trial_subs - (OLD.status = 'trial')
WHERE id = 1;

DROP TRIGGER IF EXISTS console_stats_grants_ai;
CREATE TRIGGER console_stats_grants_ai AFTER INSERT ON manual_access_grants
FOR EACH ROW UPDATE console_stats
SET manual_grants = manual_grants + (NEW.revoked_at IS NULL)
WHERE id = 1;

DROP TRIGGER IF EXISTS console_stats_grants_au;
CREATE TRIGGER console_stats_grants_au AFTER UPDATE ON manual_access_grants
FOR EACH ROW UPDATE console_stats
SET manual_grants = manual_grants + (NEW.revoked_at IS NULL) - (OLD.revoked_at IS NULL)
WHERE id = 1;

DROP TRIGGER IF EXISTS console_stats_grants_ad;
CREATE TRIGGER console_stats_grants_ad AFTER DELETE ON manual_access_grants
FOR EACH ROW UPDATE console_stats
SET manual_grants = manual_grants - (OLD.revoked_at IS NULL)
WHERE id = 1;

DROP TRIGGER IF EXISTS console_stats_addon_grants_ai;
CREATE TRIGGER console_stats_addon_grants_ai AFTER INSERT ON manual_addon_grants
FOR EACH ROW UPDATE console_stats
SET manual_addon_grants = manual_addon_grants + (NEW.revoked_at IS NULL)
WHERE id = 1;

DROP TRIGGER IF EXISTS console_stats_addon_grants_au;
CREATE TRIGGER console_stats_addon_grants_au AFTER UPDATE ON manual_addon_grants
FOR EACH ROW UPDATE console_stats
SET manual_addon_grants = manual_addon_grants + (NEW.revoked_at IS NULL) - (OLD.revoked_at IS NULL)
WHERE id = 1;

DROP TRIGGER IF EXISTS console_stats_addon_grants_ad;
CREATE TRIGGER console_stats_addon_grants_ad AFTER DELETE ON manual_addon_grants
FOR EACH ROW UPDATE console_stats
SET manual_addon_grants = manual_addon_grants - (OLD.revoked_at IS NULL)
WHERE id = 1;
