-- Newsletter subscribers are mirrored into a Resend audience by the worker job
-- websitelead.sync_newsletter_audience. These columns record the mirror state
-- so the job can backfill the existing list once and retry failures a bounded
-- number of times without re-sending anything.
ALTER TABLE voxire_newsletter_subscribers
  ADD COLUMN resend_contact_id VARCHAR(64) NULL AFTER language,
  ADD COLUMN resend_synced_at TIMESTAMP NULL AFTER resend_contact_id,
  ADD COLUMN resend_sync_attempts INT NOT NULL DEFAULT 0 AFTER resend_synced_at,
  ADD COLUMN resend_sync_error VARCHAR(255) NULL AFTER resend_sync_attempts,
  ADD INDEX idx_newsletter_resend_pending (resend_synced_at, resend_sync_attempts);
