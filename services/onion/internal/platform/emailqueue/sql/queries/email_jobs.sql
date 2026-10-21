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

-- The claim, and the only place a job leaves 'pending'.
--
-- This is a single statement, not a SELECT followed by an UPDATE, so there is
-- no window in which two workers can both decide the same row is theirs:
-- InnoDB re-evaluates the WHERE clause after it takes the row lock, so the
-- loser moves on to the next candidate and ends up affecting zero rows. The
-- caller must treat zero rows as "someone else won" and send nothing.
--
-- job_id is the claim token. A fresh UUID per claim, and it is required to be
-- NULL to be claimable, so the token identifies the winner unambiguously and
-- every later transition can be fenced on it. Nothing else writes the column.
-- name: ClaimJob :execresult
UPDATE email_jobs
SET job_id = ?, status = 'processing'
WHERE status = 'pending' AND job_id IS NULL AND scheduled_at <= NOW()
ORDER BY (priority + LEAST(TIMESTAMPDIFF(MINUTE, created_at, NOW()) DIV 60, 10)) DESC,
         scheduled_at ASC
LIMIT 1;

-- Only a row still in flight under this token is readable, so a worker cannot
-- read and send a job whose claim has already been taken away from it.
-- name: GetClaimedJob :one
SELECT id, to_email, subject, html, text_body, from_email, reply_to, attempts, max_attempts
FROM email_jobs WHERE job_id = ? AND status = 'processing';

-- Handing a claim back without spending an attempt, for a worker that is
-- shutting down mid-send rather than one that failed. Fenced on the token, so
-- releasing a claim that recovery already requeued affects nothing.
-- name: ReleaseClaim :execresult
UPDATE email_jobs
SET status = 'pending', job_id = NULL, scheduled_at = NOW()
WHERE status = 'processing' AND job_id = ?;

-- The ack. Fenced on the claim token, and execrows so the caller can see
-- whether it actually won: matching on id alone let a worker whose lease had
-- already expired stamp a job 'sent' that another worker was mid-send on, and
-- re-stamp sent_at on a repeat. Zero rows means the claim is gone, which is
-- reported rather than retried: the send already happened, so a second one
-- would only duplicate it.
-- name: MarkJobSent :execrows
UPDATE email_jobs
SET status = 'sent', sent_at = NOW(), attempts = ?, job_id = NULL
WHERE id = ? AND status = 'processing' AND job_id = ?;

-- The nack, fenced on the same token. Unfenced, a slow worker's failure arrived
-- after a takeover and put a job another worker had already sent back to
-- 'pending' (a second delivery), rewound the attempt counter written by
-- RecoverStuckJobs so a poison job could never reach max_attempts, and
-- un-dead-lettered a job that had already failed for good.
--
-- job_id is cleared by the statement rather than passed in, because the SET
-- value and the fence value cannot both be the column.
-- name: RescheduleJob :execrows
UPDATE email_jobs
SET status = ?, attempts = ?, last_error = ?, job_id = NULL,
    scheduled_at = DATE_ADD(NOW(), INTERVAL ? SECOND)
WHERE id = ? AND status = 'processing' AND job_id = ?;

-- The visibility timeout for a claim: a worker that dies or hangs leaves the
-- row in 'processing' with nothing else looking at it, so the sweeper requeues
-- it. Recovering makes the queue at-least-once, because a worker that was
-- merely slow may still deliver afterwards; that duplicate is what the fenced
-- ack and nack above bound, by refusing the stale worker's state change.
--
-- The threshold is a parameter so it is one named constant beside the send
-- timeout it has to exceed, rather than a literal here that nothing can check
-- against. Attempts increment on recovery rather than on failure, so a worker
-- that dies mid-send still counts against max_attempts.
-- name: RecoverStuckJobs :execresult
UPDATE email_jobs
SET status = CASE WHEN attempts + 1 >= max_attempts THEN 'failed' ELSE 'pending' END,
    attempts = attempts + 1,
    job_id = NULL,
    last_error = COALESCE(last_error, 'worker claim expired during delivery')
WHERE status = 'processing' AND updated_at < DATE_SUB(NOW(), INTERVAL sqlc.arg('stale_after_seconds') SECOND);

-- name: DeleteFinishedJobsOlderThan :execrows
DELETE FROM email_jobs
WHERE status IN ('sent', 'failed')
  AND created_at < ?;
