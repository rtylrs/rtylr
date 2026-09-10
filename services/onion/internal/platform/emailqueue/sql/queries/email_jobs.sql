-- name: EnqueueJob :exec
INSERT INTO email_jobs (priority, to_email, subject, html, text_body, from_email, reply_to)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- Idempotent enqueue for a caller that can name the message.
--
-- ON DUPLICATE KEY UPDATE id = id against uk_email_jobs_dedupe makes a repeat
-- a no-op rather than an error, which is what lets a retried producer, such as
-- an automation run recovered after its worker died, run again without sending
-- the recipient a second copy.
--
-- Separate statement rather than a nullable key on EnqueueJob: NULL does not
-- collide in a unique index, so a single statement would silently stop
-- deduplicating the moment a caller forgot to set one. Two statements make the
-- choice explicit at the call site.
-- name: EnqueueJobWithDedupeKey :exec
INSERT INTO email_jobs (priority, to_email, subject, html, text_body, from_email, reply_to, dedupe_key)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE id = id;

-- name: CountPendingJobs :one
SELECT COUNT(*) FROM email_jobs
WHERE status = 'pending' AND job_id IS NULL AND scheduled_at <= NOW();

-- name: ClaimJob :execresult
UPDATE email_jobs
SET job_id = ?, status = 'processing'
WHERE status = 'pending' AND job_id IS NULL AND scheduled_at <= NOW()
ORDER BY (priority + LEAST(TIMESTAMPDIFF(MINUTE, created_at, NOW()) DIV 60, 10)) DESC,
         scheduled_at ASC
LIMIT 1;

-- name: GetClaimedJob :one
SELECT id, to_email, subject, html, text_body, from_email, reply_to, attempts, max_attempts
FROM email_jobs WHERE job_id = ?;

-- name: ReleaseClaim :execresult
UPDATE email_jobs
SET status = 'pending', job_id = NULL, scheduled_at = NOW()
WHERE status = 'processing' AND job_id = ?;

-- name: MarkJobSent :exec
UPDATE email_jobs SET status = 'sent', sent_at = NOW(), attempts = ?, job_id = NULL WHERE id = ?;

-- name: RescheduleJob :exec
UPDATE email_jobs
SET status = ?, attempts = ?, last_error = ?, job_id = ?,
    scheduled_at = DATE_ADD(NOW(), INTERVAL ? SECOND)
WHERE id = ?;

-- name: RecoverStuckJobs :execresult
UPDATE email_jobs
SET status = CASE WHEN attempts + 1 >= max_attempts THEN 'failed' ELSE 'pending' END,
    attempts = attempts + 1,
    job_id = NULL,
    last_error = COALESCE(last_error, 'worker claim expired during delivery')
WHERE status = 'processing' AND updated_at < DATE_SUB(NOW(), INTERVAL 5 MINUTE);

-- name: DeleteFinishedJobsOlderThan :execrows
DELETE FROM email_jobs
WHERE status IN ('sent', 'failed')
  AND created_at < ?;
