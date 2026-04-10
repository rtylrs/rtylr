-- name: ListDocuments :many
SELECT id, owner_id, staff_id, name, type, file_url, expiry_date, created_at
FROM hr_documents
WHERE owner_id = ?
ORDER BY created_at DESC;

-- name: CreateDocument :execresult
INSERT INTO hr_documents (owner_id, staff_id, name, type, file_url, expiry_date)
VALUES (?, ?, ?, ?, ?, ?);

-- name: DeleteDocument :exec
DELETE FROM hr_documents WHERE id = ? AND owner_id = ?;
