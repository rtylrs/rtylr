-- Newsletter subscribers are mirrored into a Resend audience by the worker job
-- websitelead.sync_newsletter_audience. These columns record the mirror state
-- so the job can backfill the existing list once and retry failures a bounded
-- number of times without re-sending anything.
SET @newsletter_resend_missing := (
    SELECT COUNT(*) = 0
    FROM information_schema.columns
    WHERE table_schema = DATABASE()
      AND table_name = 'voxire_newsletter_subscribers'
      AND column_name = 'resend_contact_id'
);
SET @newsletter_resend_sql := IF(
    @newsletter_resend_missing,
    'ALTER TABLE voxire_newsletter_subscribers
       ADD COLUMN resend_contact_id VARCHAR(64) NULL AFTER language,
       ADD COLUMN resend_synced_at TIMESTAMP NULL AFTER resend_contact_id,
       ADD COLUMN resend_sync_attempts INT NOT NULL DEFAULT 0 AFTER resend_synced_at,
       ADD COLUMN resend_sync_error VARCHAR(255) NULL AFTER resend_sync_attempts,
       ADD INDEX idx_newsletter_resend_pending (resend_synced_at, resend_sync_attempts)',
    'SELECT 1'
);
PREPARE newsletter_resend_stmt FROM @newsletter_resend_sql;
EXECUTE newsletter_resend_stmt;
DEALLOCATE PREPARE newsletter_resend_stmt;
