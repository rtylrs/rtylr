-- name: CreateDocument :execresult
INSERT INTO hr_employee_documents (owner_id, employee_id, document_type, title, storage_mode, file_url, external_url, uploaded_by)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetDocument :one
SELECT id, owner_id, employee_id, document_type, title, storage_mode, file_url, external_url, expiry_date, uploaded_by, created_at, updated_at
FROM hr_employee_documents
WHERE owner_id = ? AND id = ?;

-- name: ListDocuments :many
SELECT id, owner_id, employee_id, document_type, title, storage_mode, file_url, external_url, expiry_date, uploaded_by, created_at, updated_at
FROM hr_employee_documents
WHERE owner_id = ? AND employee_id = ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: CountDocuments :one
SELECT COUNT(*) AS total
FROM hr_employee_documents
WHERE owner_id = ? AND employee_id = ?;

-- name: DeleteDocument :exec
DELETE FROM hr_employee_documents
WHERE owner_id = ? AND id = ?;
