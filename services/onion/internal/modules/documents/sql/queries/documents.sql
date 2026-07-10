-- name: ListDocuments :many
SELECT id, owner_id, business_id, staff_id, name, type, file_url, expiry_date, created_at
FROM hr_documents
WHERE business_id = ?
ORDER BY created_at DESC;

-- name: CreateDocument :execresult
INSERT INTO hr_documents (owner_id, business_id, staff_id, name, type, file_url, expiry_date)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: DeleteDocument :exec
DELETE FROM hr_documents WHERE id = ? AND business_id = ?;

-- name: ListDocumentsPaginated :many
SELECT id, owner_id, business_id, staff_id, name, type, file_url, expiry_date, created_at
FROM hr_documents
WHERE business_id = ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: CountDocuments :one
SELECT COUNT(*) FROM hr_documents WHERE business_id = ?;
