-- An opt-in idempotency key for queued email.
--
-- Added so an abandoned automation run can be retried safely. The automation
-- executor claims a run before performing its actions, so a worker that dies
-- mid-run leaves the row stranded; onion#1190 recovers those, but had to
-- dead-letter the automation ones rather than requeue them, because re-running
-- could send a customer a second copy of a message they already have and the
-- recipient of an email has nothing to dedupe on. This gives the enqueue side
-- the key it needs, so the retry becomes safe and the run can be requeued like
-- a webhook delivery.
--
-- NULL means no deduplication, which is every existing row and every caller
-- that does not opt in. MySQL permits any number of NULLs in a unique index,
-- so nothing already queued or any other producer is affected.
--
-- Named as its own zz_ file rather than an edit to shared/email_jobs.sql
-- because migrate.sh records applied files by name: editing an already-applied
-- file is a silent no-op. Guarded on INFORMATION_SCHEMA because zz_ files
-- re-run on every pass, and migrate.sh records a FAILED migration as applied
-- behind a green check, so an unguarded ALTER would fail silently and forever
-- after the first run.
SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
   WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'email_jobs'
     AND COLUMN_NAME = 'dedupe_key') = 0,
  'ALTER TABLE email_jobs ADD COLUMN dedupe_key VARCHAR(191) NULL DEFAULT NULL AFTER job_id',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 191 rather than 255 so the index fits utf8mb4 under the older InnoDB row
-- formats, matching platform_events.dedupe_key.
SET @idx := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
   WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'email_jobs'
     AND INDEX_NAME = 'uk_email_jobs_dedupe') = 0,
  'ALTER TABLE email_jobs ADD UNIQUE KEY uk_email_jobs_dedupe (dedupe_key)',
  'SELECT 1'
);
PREPARE stmt2 FROM @idx; EXECUTE stmt2; DEALLOCATE PREPARE stmt2;
