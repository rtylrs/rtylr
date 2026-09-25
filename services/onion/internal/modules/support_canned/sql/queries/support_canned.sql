-- name: ListCannedResponses :many
SELECT id, business_id, name, body, category, shortcut, created_at
FROM support_canned_responses WHERE business_id = ?
ORDER BY name ASC
LIMIT 200;

-- name: CreateCannedResponse :execresult
INSERT INTO support_canned_responses (business_id, name, body, category, shortcut)
VALUES (?, ?, ?, ?, ?);

-- name: UpdateCannedResponse :exec
UPDATE support_canned_responses SET name = ?, body = ?, category = ?, shortcut = ?
WHERE business_id = ? AND id = ?;

-- name: DeleteCannedResponse :exec
DELETE FROM support_canned_responses WHERE business_id = ? AND id = ?;
