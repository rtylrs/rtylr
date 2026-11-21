-- name: CountPushSubscriptions :one
SELECT COUNT(*)
FROM pager_push_subscriptions
WHERE business_id = ? AND ticket_id = ?;

-- name: SavePushSubscription :exec
INSERT INTO pager_push_subscriptions (business_id, ticket_id, endpoint, endpoint_hash, p256dh, auth, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE p256dh = VALUES(p256dh), auth = VALUES(auth);

-- name: ListPushSubscriptions :many
SELECT id, business_id, ticket_id, endpoint, p256dh, auth
FROM pager_push_subscriptions
WHERE business_id = ? AND ticket_id = ?
ORDER BY id;

-- name: DeletePushSubscription :exec
DELETE FROM pager_push_subscriptions
WHERE id = ? AND business_id = ?;

-- name: DeleteTicketPushSubscriptions :exec
DELETE FROM pager_push_subscriptions
WHERE business_id = ? AND ticket_id = ?;

-- name: DeleteStalePushSubscriptions :exec
DELETE FROM pager_push_subscriptions
WHERE created_at < sqlc.arg(open_since)
LIMIT 500;
