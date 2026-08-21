-- Profile photo and bio for a staff member, added for tryrtylr/rtylr#875.
--
-- Both are optional and unset by default: a staff member with no photo falls
-- back to the hub account avatar, which is what Flow's profile screen already
-- does. Declared in hr.staff.v1 by tryrtylr/voxie#38.
--
-- `bio` is deliberately its own column rather than reusing `notes`. `notes` is
-- HR-admin commentary about a person; rendering it as an "About" section would
-- publish things like "performance concerns, see Q3 review" to anyone who can
-- read the profile. A dead write path in the HR app mapped bio onto notes, and
-- that mapping was the bug.
--
-- Guarded rather than bare. migrate.sh runs every file in this directory on
-- every invocation and passes --force, so a plain ALTER would raise a duplicate
-- column error on the second run. That error is counted and then reported as
-- "✅ Migrations complete with N warning(s)", which means a real failure looks
-- exactly like this one. Guarding keeps the warning count meaningful.

SET @add_photo_url := IF(
    (SELECT COUNT(*) FROM information_schema.columns
     WHERE table_schema = DATABASE()
       AND table_name = 'hr_staff'
       AND column_name = 'photo_url') = 0,
    'ALTER TABLE hr_staff ADD COLUMN photo_url VARCHAR(512) NULL AFTER notes',
    'DO 0');
PREPARE add_photo_url FROM @add_photo_url;
EXECUTE add_photo_url;
DEALLOCATE PREPARE add_photo_url;

SET @add_bio := IF(
    (SELECT COUNT(*) FROM information_schema.columns
     WHERE table_schema = DATABASE()
       AND table_name = 'hr_staff'
       AND column_name = 'bio') = 0,
    'ALTER TABLE hr_staff ADD COLUMN bio TEXT NULL AFTER photo_url',
    'DO 0');
PREPARE add_bio FROM @add_bio;
EXECUTE add_bio;
DEALLOCATE PREPARE add_bio;
