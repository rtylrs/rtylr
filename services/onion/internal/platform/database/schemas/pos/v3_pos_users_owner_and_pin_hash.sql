-- ═══════════════════════════════════════════════════════════════════
-- pos_users security overhaul
--   1. Rename tenancy key column business_id → owner_id to match every
--      other tenant-scoped table in the schema.
--   2. Add nullable pin_hash column for bcrypt-hashed PINs. Existing rows
--      keep their plaintext pin until the first successful verify-pin
--      call migrates them (see "dual-path verify" comment in the
--      repository).
--   3. Drop the old UNIQUE(business_id, pin) — bcrypt hashes are unique
--      per row by design, so the uniqueness guarantee no longer fits.
--      Uniqueness of plaintext PINs within a business is enforced at the
--      service layer on create/update instead.
--
-- This migration is idempotent: re-running it on an already-migrated DB
-- is a no-op.
-- ═══════════════════════════════════════════════════════════════════

-- 1. Rename business_id → owner_id. MySQL 8's RENAME COLUMN is
--    non-idempotent on its own, so we guard with information_schema.
SET @has_business_id := (
    SELECT COUNT(*) FROM information_schema.columns
    WHERE table_schema = DATABASE()
      AND table_name   = 'pos_users'
      AND column_name  = 'business_id'
);
SET @has_owner_id := (
    SELECT COUNT(*) FROM information_schema.columns
    WHERE table_schema = DATABASE()
      AND table_name   = 'pos_users'
      AND column_name  = 'owner_id'
);
SET @stmt := IF(
    @has_business_id = 1 AND @has_owner_id = 0,
    'ALTER TABLE pos_users RENAME COLUMN business_id TO owner_id',
    'SELECT 1'
);
PREPARE s FROM @stmt;
EXECUTE s;
DEALLOCATE PREPARE s;

-- 2. Add pin_hash if missing. IF NOT EXISTS is supported on MySQL 8.
ALTER TABLE pos_users
    ADD COLUMN IF NOT EXISTS pin_hash VARCHAR(255) DEFAULT NULL;

-- 3. Plaintext pin column must become nullable/defaulted so the app can
--    clear it post-migration without breaking the NOT NULL constraint.
ALTER TABLE pos_users
    MODIFY COLUMN pin VARCHAR(10) NOT NULL DEFAULT '';

-- 4. Replace the business-id-scoped index with an owner-id-scoped one.
SET @has_old_idx := (
    SELECT COUNT(*) FROM information_schema.statistics
    WHERE table_schema = DATABASE()
      AND table_name   = 'pos_users'
      AND index_name   = 'idx_pos_users_business'
);
SET @stmt := IF(@has_old_idx = 1,
    'ALTER TABLE pos_users DROP INDEX idx_pos_users_business',
    'SELECT 1'
);
PREPARE s FROM @stmt;
EXECUTE s;
DEALLOCATE PREPARE s;

SET @has_new_idx := (
    SELECT COUNT(*) FROM information_schema.statistics
    WHERE table_schema = DATABASE()
      AND table_name   = 'pos_users'
      AND index_name   = 'idx_pos_users_owner'
);
SET @stmt := IF(@has_new_idx = 0,
    'ALTER TABLE pos_users ADD INDEX idx_pos_users_owner (owner_id)',
    'SELECT 1'
);
PREPARE s FROM @stmt;
EXECUTE s;
DEALLOCATE PREPARE s;

-- 5. Drop the legacy UNIQUE(business_id, pin) constraint. bcrypt hashes
--    are unique per row so the constraint is no longer applicable, and
--    the rename above would have left it referencing the renamed column
--    anyway.
SET @has_unique := (
    SELECT COUNT(*) FROM information_schema.statistics
    WHERE table_schema = DATABASE()
      AND table_name   = 'pos_users'
      AND index_name   = 'unique_pin_business'
);
SET @stmt := IF(@has_unique >= 1,
    'ALTER TABLE pos_users DROP INDEX unique_pin_business',
    'SELECT 1'
);
PREPARE s FROM @stmt;
EXECUTE s;
DEALLOCATE PREPARE s;
