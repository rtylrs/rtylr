-- name: CreateDocument :execresult
INSERT INTO hr_employee_documents (owner_id, business_id, employee_id, document_type, title, storage_mode, file_url, external_url, expiry_date, uploaded_by)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetDocument :one
SELECT id, owner_id, business_id, employee_id, document_type, title, storage_mode, file_url, external_url, expiry_date, uploaded_by, created_at, updated_at
FROM hr_employee_documents
WHERE business_id = ? AND id = ?;

-- name: ListDocuments :many
SELECT id, owner_id, business_id, employee_id, document_type, title, storage_mode, file_url, external_url, expiry_date, uploaded_by, created_at, updated_at
FROM hr_employee_documents
WHERE business_id = sqlc.arg(business_id)
  AND (sqlc.arg(employee_id) = 0 OR employee_id = sqlc.arg(employee_id))
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: CountDocuments :one
SELECT COUNT(*) AS total
FROM hr_employee_documents
WHERE business_id = sqlc.arg(business_id)
  AND (sqlc.arg(employee_id) = 0 OR employee_id = sqlc.arg(employee_id));

-- name: DeleteDocument :execresult
DELETE FROM hr_employee_documents
WHERE business_id = ? AND id = ?;

-- name: CountExpiringDocuments :one
SELECT COUNT(*) AS total
FROM hr_employee_documents
WHERE business_id = ? AND expiry_date IS NOT NULL AND expiry_date <= ?;

-- name: ListExpiringDocuments :many
SELECT id, owner_id, business_id, employee_id, document_type, title, storage_mode, file_url, external_url, expiry_date, uploaded_by, created_at, updated_at
FROM hr_employee_documents
WHERE business_id = ? AND expiry_date IS NOT NULL AND expiry_date <= ?
ORDER BY expiry_date ASC
LIMIT ? OFFSET ?;
