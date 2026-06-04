-- name: EnqueueJob :exec
INSERT INTO email_jobs (priority, to_email, subject, html, text_body, from_email, reply_to)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: CountPendingJobs :one
SELECT COUNT(*) FROM email_jobs
WHERE status = 'pending' AND job_id IS NULL AND scheduled_at <= NOW();

-- name: ClaimJob :execresult
UPDATE email_jobs
SET job_id = ?, status = 'processing'
WHERE status = 'pending' AND job_id IS NULL AND scheduled_at <= NOW()
ORDER BY priority DESC, scheduled_at ASC
LIMIT 1;

-- name: GetClaimedJob :one
SELECT id, to_email, subject, html, text_body, from_email, reply_to, attempts, max_attempts
FROM email_jobs WHERE job_id = ?;

-- name: MarkJobSent :exec
UPDATE email_jobs SET status = 'sent', sent_at = NOW(), attempts = ?, job_id = NULL WHERE id = ?;

-- name: RescheduleJob :exec
UPDATE email_jobs
SET status = ?, attempts = ?, last_error = ?, job_id = ?,
    scheduled_at = DATE_ADD(NOW(), INTERVAL ? SECOND)
WHERE id = ?;

-- name: RecoverStuckJobs :execresult
UPDATE email_jobs
SET status = 'pending', job_id = NULL
WHERE status = 'processing' AND updated_at < DATE_SUB(NOW(), INTERVAL 5 MINUTE);
