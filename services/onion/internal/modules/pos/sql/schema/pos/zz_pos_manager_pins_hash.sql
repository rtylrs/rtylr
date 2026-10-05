-- Manager PINs are bcrypt hashes, never plaintext.
--
-- The legacy `pin` column held the PIN in cleartext VARCHAR(10). A dump, a
-- backup, a replica or read-only SQL access on any of them yielded every
-- manager PIN in the tenant immediately, and a brute-force lockout does nothing
-- about that: the credential is already known, so the attacker never has to
-- guess. This mirrors schemas/pos/zz_pos_users_remove_plaintext_pin.sql for the
-- pos_users table.
--
-- Step 1 is additive and idempotent so the code can be deployed against either
-- shape. Rows already present keep their `pin` value in step 1; step 3 drops it.

SET @sql := IF(
    EXISTS (
        SELECT 1
        FROM INFORMATION_SCHEMA.COLUMNS
        WHERE TABLE_SCHEMA = DATABASE()
          AND TABLE_NAME = 'pos_manager_pins'
          AND COLUMN_NAME = 'pin_hash'
    ),
    'SELECT 1',
    'ALTER TABLE pos_manager_pins ADD COLUMN pin_hash VARCHAR(255) DEFAULT NULL'
);
PREPARE add_manager_pin_hash FROM @sql;
EXECUTE add_manager_pin_hash;
DEALLOCATE PREPARE add_manager_pin_hash;

-- Step 2: no legacy PIN can be re-hashed in place. bcrypt is one-way, so a
-- plaintext row can only be converted by reading the plaintext and re-hashing
-- it, which is an application-level operation, not a SQL one. The rows are left
-- with a NULL pin_hash, which the verify path treats as "no usable credential"
-- and refuses. That is the forced reset: every pre-existing manager PIN stops
-- working on the next release.

-- Step 3: drop the plaintext column. The verify path no longer reads it, but
-- leaving it would keep every historical manager PIN readable in every backup
-- taken from now on, which is the vulnerability being closed here.
SET @sql := IF(
    EXISTS (
        SELECT 1
        FROM INFORMATION_SCHEMA.COLUMNS
        WHERE TABLE_SCHEMA = DATABASE()
          AND TABLE_NAME = 'pos_manager_pins'
          AND COLUMN_NAME = 'pin'
    ),
    'ALTER TABLE pos_manager_pins DROP COLUMN pin',
    'SELECT 1'
);
PREPARE drop_legacy_manager_pin FROM @sql;
EXECUTE drop_legacy_manager_pin;
DEALLOCATE PREPARE drop_legacy_manager_pin;

-- OPERATOR ACTION REQUIRED, and it is not optional:
--   Every pre-existing manager PIN is invalid after this migration. There is no
--   seamless conversion and no in-place re-hash; a forced reset of every
--   manager PIN at every business is the honest answer.
--
-- The reset no longer requires out-of-band seeding. pos_manager now exposes
-- POST /pos/manager/v1/set-pin, which stores a cost-12 bcrypt hash of a 6-12
-- digit PIN. It is restricted to the business owner and additionally requires
-- that owner to re-enter their account password, so a stolen or left-open
-- session cannot mint an override credential.
--
-- Until each manager is re-provisioned through that route they have no usable
-- manager credential and the override is refused, which is the intended
-- fail-closed outcome rather than a silent fall back to the old PIN.
