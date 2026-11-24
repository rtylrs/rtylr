-- open_since is the creation time before which a labelled ticket no longer
-- counts as open, unlabelled_since the same for a ticket nobody labelled.

-- name: GetSettings :one
SELECT enabled, label_mode, review_url, prep_minutes, customer_numbers, ringtone
FROM pager_settings
WHERE business_id = ?;

-- name: UpsertSettings :exec
INSERT INTO pager_settings (business_id, enabled, label_mode, review_url, prep_minutes, customer_numbers, ringtone)
VALUES (?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE enabled = VALUES(enabled), label_mode = VALUES(label_mode),
  review_url = VALUES(review_url), prep_minutes = VALUES(prep_minutes),
  customer_numbers = VALUES(customer_numbers), ringtone = VALUES(ringtone);

-- name: CreateTicket :execresult
INSERT INTO pager_tickets (business_id, match_number, status, source, public_token, phone_linked, created_at)
VALUES (?, ?, 'waiting', 'customer', ?, TRUE, ?);

-- A customer who typed their receipt number: labelled from the start.
-- uk_pager_tickets_open_number refuses a number that already has an open ticket.
-- name: CreateNumberTicket :execresult
INSERT INTO pager_tickets (business_id, number, label, label_kind, status, source, public_token, phone_linked, created_at, expected_ready_at)
VALUES (?, ?, ?, 'number', 'waiting', 'customer', ?, TRUE, ?, ?);

-- name: CountOpenTickets :one
SELECT COUNT(*)
FROM pager_tickets
WHERE business_id = ? AND status IN ('waiting', 'ready') AND created_at >= sqlc.arg(open_since)
  AND (label_kind IS NOT NULL OR created_at >= sqlc.arg(unlabelled_since));

-- name: ListOpenMatchCodes :many
SELECT COALESCE(match_number, match_word, match_code, '') AS code
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

-- name: ListPagerBusinessIDs :many
SELECT business_id
FROM pager_settings;

-- name: ListStaleLabelledTicketIDs :many
SELECT id
FROM pager_tickets
WHERE business_id = ? AND status IN ('waiting', 'ready') AND label_kind IS NOT NULL
  AND created_at < sqlc.arg(open_since);

-- name: ExpireStaleLabelledTicket :exec
UPDATE pager_tickets
SET status = 'expired'
WHERE id = ? AND business_id = ? AND status IN ('waiting', 'ready') AND label_kind IS NOT NULL
  AND created_at < sqlc.arg(open_since);

-- name: ExpireStaleTicket :exec
UPDATE pager_tickets
SET status = 'expired'
WHERE business_id = ? AND open_number = ? AND created_at < sqlc.arg(open_since);

-- name: ListOpenTickets :many
SELECT id, business_id, label, label_kind, match_code, match_word, match_number, status, public_token,
       created_at, ready_at, reminded_at, acknowledged_at, expected_ready_at
FROM pager_tickets
WHERE business_id = ? AND status IN ('waiting', 'ready') AND created_at >= sqlc.arg(open_since)
  AND (label_kind IS NOT NULL OR created_at >= sqlc.arg(unlabelled_since))
ORDER BY id;

-- name: GetTicket :one
SELECT id, business_id, label, label_kind, match_code, match_word, match_number, status, public_token,
       created_at, ready_at, reminded_at, acknowledged_at, expected_ready_at
FROM pager_tickets
WHERE id = ? AND business_id = ?;

-- name: GetTicketByToken :one
SELECT id, business_id, label, label_kind, match_code, match_word, match_number, status, public_token,
       created_at, ready_at, reminded_at, acknowledged_at, expected_ready_at
FROM pager_tickets
WHERE public_token = ?;

-- name: CountWaitingAhead :one
SELECT COUNT(*)
FROM pager_tickets
WHERE business_id = ? AND status = 'waiting' AND label_kind IS NOT NULL
  AND created_at >= sqlc.arg(open_since) AND id < ?;

-- name: LabelTicket :exec
UPDATE pager_tickets
SET label = ?, label_kind = ?, number = ?, expected_ready_at = ?
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

-- Bounded so a venue's busiest month cannot load an unbounded number of rows.
-- name: ListLabelledTicketTimes :many
SELECT created_at, ready_at
FROM pager_tickets
WHERE business_id = ? AND label_kind IS NOT NULL AND created_at >= sqlc.arg(since)
ORDER BY id
LIMIT 50000;
