-- A user's chosen avatar (photo, illustrated, initials colour or emoji), as
-- JSON. NULL keeps the default: initials on a colour derived from the user id.
-- Additive and guarded, so a stopped run can be retried.
SET @users_avatar_column_exists := (
    SELECT COUNT(*) > 0
    FROM information_schema.columns
    WHERE table_schema = DATABASE()
      AND table_name = 'users'
      AND column_name = 'avatar'
);
SET @users_add_avatar_column_sql := IF(
    @users_avatar_column_exists,
    'SELECT 1',
    'ALTER TABLE users ADD COLUMN avatar varchar(1024) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT ''chosen avatar as JSON; NULL = default initials'' AFTER two_factor_enabled'
);
PREPARE users_add_avatar_column FROM @users_add_avatar_column_sql;
EXECUTE users_add_avatar_column;
DEALLOCATE PREPARE users_add_avatar_column;
