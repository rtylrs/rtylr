-- name: ListDocuments :many
SELECT id, owner_id, business_id, staff_id, name, type, file_url, expiry_date, created_at
FROM hr_documents
WHERE business_id = ?
ORDER BY created_at DESC LIMIT 100;

-- name: CreateDocument :execresult
INSERT INTO hr_documents (owner_id, business_id, name, type, file_url, expiry_date)
VALUES (?, ?, ?, ?, ?, ?);

-- name: CreateDocumentForStaff :execresult
INSERT INTO hr_documents (owner_id, business_id, staff_id, name, type, file_url, expiry_date)
SELECT sqlc.arg(owner_id), sqlc.arg(business_id), staff.id,
       sqlc.arg(name), sqlc.arg(type), sqlc.arg(file_url), sqlc.arg(expiry_date)
FROM hr_staff AS staff
WHERE staff.id = sqlc.arg(staff_id)
  AND staff.business_id = sqlc.arg(staff_business_id)
LIMIT 1;

-- name: DeleteDocument :execresult
DELETE FROM hr_documents WHERE id = ? AND business_id = ?;

-- name: ListDocumentsPaginated :many
SELECT id, owner_id, business_id, staff_id, name, type, file_url, expiry_date, created_at
FROM hr_documents
WHERE business_id = ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: CountDocuments :one
SELECT COUNT(*) FROM hr_documents WHERE business_id = ?;
