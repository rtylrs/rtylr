-- name: CreateImportBatch :execresult
INSERT INTO crm_import_batches (business_id, created_by, file_name, total_rows) VALUES (?, ?, ?, ?);

-- name: GetImportBatch :one
SELECT * FROM crm_import_batches WHERE id = ? AND business_id = ?;

-- name: GetImportBatchForUpdate :one
SELECT * FROM crm_import_batches WHERE id = ? AND business_id = ? FOR UPDATE;

-- name: ListImportBatches :many
SELECT * FROM crm_import_batches WHERE business_id = ? ORDER BY created_at DESC, id DESC LIMIT 20;

-- name: AddImportBatchCounts :exec
UPDATE crm_import_batches
SET created_count = created_count + ?, updated_count = updated_count + ?,
    skipped_count = skipped_count + ?, failed_count = failed_count + ?
WHERE id = ? AND business_id = ?;

-- name: SetImportBatchStatus :exec
UPDATE crm_import_batches SET status = ?, completed_at = ? WHERE id = ? AND business_id = ?;
