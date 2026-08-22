-- sqlc-only schema supplement for Flow's runtime migration.
--
-- The production migration uses guarded dynamic SQL because MySQL 8.4 does
-- not support ADD COLUMN IF NOT EXISTS. sqlc does not evaluate prepared SQL,
-- so keep the resulting columns explicit here without placing this file in a
-- runtime migration directory.
ALTER TABLE hr_time_entries
    ADD COLUMN work_day_id BIGINT NULL,
    ADD COLUMN work_type_id BIGINT NULL,
    ADD COLUMN work_mode VARCHAR(20) NULL,
    ADD COLUMN timezone VARCHAR(64) NULL,
    ADD COLUMN client_key VARCHAR(64) NULL,
    ADD COLUMN logged_by BIGINT NULL;
