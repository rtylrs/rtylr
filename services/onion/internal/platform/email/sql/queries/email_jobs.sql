-- name: EnqueueEmail :exec
INSERT INTO email_jobs (priority, to_email, subject, html, from_email, reply_to)
VALUES (?, ?, ?, ?, ?, ?);
