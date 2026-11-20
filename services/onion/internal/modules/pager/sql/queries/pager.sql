-- name: GetSettings :one
SELECT enabled, strict
FROM pager_settings
WHERE business_id = ?;

-- name: UpsertSettings :exec
INSERT INTO pager_settings (business_id, enabled, strict)
VALUES (?, ?, ?)
ON DUPLICATE KEY UPDATE enabled = VALUES(enabled), strict = VALUES(strict);

-- name: CreateTicket :execresult
INSERT INTO pager_tickets (business_id, number, status, source, public_token, phone_linked, created_at)
VALUES (?, ?, 'waiting', ?, ?, ?, ?);

-- name: GetOpenTicketByNumber :one
SELECT id, business_id, number, status, source, public_token, phone_linked,
       created_at, ready_at, reminded_at, acknowledged_at
FROM pager_tickets
WHERE business_id = ? AND open_number = ? AND created_at >= sqlc.arg(open_since);

-- name: ExpireStaleTicket :exec
UPDATE pager_tickets
SET status = 'expired'
WHERE business_id = ? AND open_number = ? AND created_at < sqlc.arg(open_since);

-- name: CountOpenCustomerTickets :one
SELECT COUNT(*)
FROM pager_tickets
WHERE business_id = ? AND status IN ('waiting', 'ready') AND source = 'customer'
  AND created_at >= sqlc.arg(open_since);

-- name: ListOpenTickets :many
SELECT id, business_id, number, status, source, public_token, phone_linked,
       created_at, ready_at, reminded_at, acknowledged_at
FROM pager_tickets
WHERE business_id = ? AND status IN ('waiting', 'ready') AND created_at >= sqlc.arg(open_since)
ORDER BY id;

-- name: GetTicket :one
SELECT id, business_id, number, status, source, public_token, phone_linked,
       created_at, ready_at, reminded_at, acknowledged_at
FROM pager_tickets
WHERE id = ? AND business_id = ?;

-- name: GetTicketByToken :one
SELECT id, business_id, number, status, source, public_token, phone_linked,
       created_at, ready_at, reminded_at, acknowledged_at
FROM pager_tickets
WHERE public_token = ?;

-- name: CountWaitingAhead :one
SELECT COUNT(*)
FROM pager_tickets
WHERE business_id = ? AND status = 'waiting' AND created_at >= sqlc.arg(open_since) AND id < ?;

-- name: LinkPhone :exec
UPDATE pager_tickets
SET phone_linked = TRUE
WHERE id = ? AND business_id = ?;

-- name: MarkTicketReady :exec
UPDATE pager_tickets
SET status = 'ready', ready_at = ?, reminded_at = NULL, acknowledged_at = NULL
WHERE id = ? AND business_id = ? AND status = 'waiting' AND created_at >= sqlc.arg(open_since);

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
WHERE id = ? AND business_id = ? AND status IN ('waiting', 'ready') AND created_at >= sqlc.arg(open_since);

-- name: AcknowledgeTicket :exec
UPDATE pager_tickets
SET acknowledged_at = ?
WHERE id = ? AND business_id = ? AND status = 'ready' AND acknowledged_at IS NULL
  AND created_at >= sqlc.arg(open_since);
