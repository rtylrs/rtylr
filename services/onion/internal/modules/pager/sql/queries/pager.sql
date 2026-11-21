-- open_since is the creation time before which a labelled ticket no longer
-- counts as open, unlabelled_since the same for a ticket nobody labelled.

-- name: GetSettings :one
SELECT enabled, label_mode
FROM pager_settings
WHERE business_id = ?;

-- name: UpsertSettings :exec
INSERT INTO pager_settings (business_id, enabled, label_mode)
VALUES (?, ?, ?)
ON DUPLICATE KEY UPDATE enabled = VALUES(enabled), label_mode = VALUES(label_mode);

-- name: CreateTicket :execresult
INSERT INTO pager_tickets (business_id, match_word, status, source, public_token, phone_linked, created_at)
VALUES (?, ?, 'waiting', 'customer', ?, TRUE, ?);

-- name: ListOpenMatchCodes :many
SELECT COALESCE(match_word, match_code, '') AS code
FROM pager_tickets
WHERE business_id = ? AND status IN ('waiting', 'ready') AND label_kind IS NULL
  AND created_at >= sqlc.arg(unlabelled_since);

-- Expiring goes through a plain read first: an UPDATE over the venue's range
-- of open tickets would lock it against the scans inserting into it, and a
-- rush of scans then deadlocks.
-- name: ListStaleUnlabelledTicketIDs :many
SELECT id
FROM pager_tickets
WHERE business_id = ? AND status IN ('waiting', 'ready') AND label_kind IS NULL
  AND created_at < sqlc.arg(unlabelled_since);

-- name: ExpireStaleUnlabelledTicket :exec
UPDATE pager_tickets
SET status = 'expired'
WHERE id = ? AND business_id = ? AND status IN ('waiting', 'ready') AND label_kind IS NULL
  AND created_at < sqlc.arg(unlabelled_since);

-- name: ExpireStaleTicket :exec
UPDATE pager_tickets
SET status = 'expired'
WHERE business_id = ? AND open_number = ? AND created_at < sqlc.arg(open_since);

-- name: ListOpenTickets :many
SELECT id, business_id, label, label_kind, match_code, match_word, status, public_token,
       created_at, ready_at, reminded_at, acknowledged_at
FROM pager_tickets
WHERE business_id = ? AND status IN ('waiting', 'ready') AND created_at >= sqlc.arg(open_since)
  AND (label_kind IS NOT NULL OR created_at >= sqlc.arg(unlabelled_since))
ORDER BY id;

-- name: GetTicket :one
SELECT id, business_id, label, label_kind, match_code, match_word, status, public_token,
       created_at, ready_at, reminded_at, acknowledged_at
FROM pager_tickets
WHERE id = ? AND business_id = ?;

-- name: GetTicketByToken :one
SELECT id, business_id, label, label_kind, match_code, match_word, status, public_token,
       created_at, ready_at, reminded_at, acknowledged_at
FROM pager_tickets
WHERE public_token = ?;

-- name: CountWaitingAhead :one
SELECT COUNT(*)
FROM pager_tickets
WHERE business_id = ? AND status = 'waiting' AND label_kind IS NOT NULL
  AND created_at >= sqlc.arg(open_since) AND id < ?;

-- name: LabelTicket :exec
UPDATE pager_tickets
SET label = ?, label_kind = ?, number = ?
WHERE id = ? AND business_id = ? AND status = 'waiting' AND created_at >= sqlc.arg(open_since)
  AND (label_kind IS NOT NULL OR created_at >= sqlc.arg(unlabelled_since));

-- name: MarkTicketReady :exec
UPDATE pager_tickets
SET status = 'ready', ready_at = ?, reminded_at = NULL, acknowledged_at = NULL
WHERE id = ? AND business_id = ? AND status = 'waiting' AND label_kind IS NOT NULL
  AND created_at >= sqlc.arg(open_since);

-- name: RemindTicket :exec
UPDATE pager_tickets
SET reminded_at = ?, acknowledged_at = NULL
WHERE id = ? AND business_id = ? AND status = 'ready' AND created_at >= sqlc.arg(open_since);

-- name: UndoTicketReady :exec
UPDATE pager_tickets
SET status = 'waiting', ready_at = NULL, reminded_at = NULL, acknowledged_at = NULL
WHERE id = ? AND business_id = ? AND status = 'ready' AND created_at >= sqlc.arg(open_since);

-- name: CollectTicket :exec
UPDATE pager_tickets
SET status = 'collected'
WHERE id = ? AND business_id = ? AND status IN ('waiting', 'ready') AND created_at >= sqlc.arg(open_since)
  AND (label_kind IS NOT NULL OR created_at >= sqlc.arg(unlabelled_since));

-- name: AcknowledgeTicket :exec
UPDATE pager_tickets
SET acknowledged_at = ?
WHERE id = ? AND business_id = ? AND status = 'ready' AND acknowledged_at IS NULL
  AND created_at >= sqlc.arg(open_since);
