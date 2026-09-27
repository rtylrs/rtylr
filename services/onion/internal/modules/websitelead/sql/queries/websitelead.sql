-- name: CreateWebsiteLead :execresult
INSERT INTO voxire_website_leads (name, email, phone, company, service, project, budget, source)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: CreateNewsletterSubscriber :execresult
INSERT INTO voxire_newsletter_subscribers (email, source, post_slug, post_title, post_category, language)
VALUES (?, ?, ?, ?, ?, ?);

-- name: ListUnsyncedNewsletterSubscribers :many
SELECT id, email
FROM voxire_newsletter_subscribers
WHERE resend_synced_at IS NULL AND resend_sync_attempts < ?
ORDER BY id
LIMIT ?;

-- name: MarkNewsletterSubscriberSynced :exec
UPDATE voxire_newsletter_subscribers
SET resend_contact_id = ?, resend_synced_at = CURRENT_TIMESTAMP, resend_sync_error = NULL
WHERE id = ?;

-- name: MarkNewsletterSubscriberSyncFailed :exec
UPDATE voxire_newsletter_subscribers
SET resend_sync_attempts = resend_sync_attempts + 1, resend_sync_error = ?
WHERE id = ?;
