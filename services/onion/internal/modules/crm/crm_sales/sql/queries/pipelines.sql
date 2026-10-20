-- name: ListPipelines :many
SELECT * FROM crm_pipelines WHERE business_id = ? ORDER BY position, id;

-- name: ListPipelineStages :many
SELECT * FROM crm_pipeline_stages WHERE business_id = ? ORDER BY pipeline_id, position, id;

-- name: GetPipeline :one
SELECT * FROM crm_pipelines WHERE id = ? AND business_id = ?;

-- name: GetDefaultPipeline :one
SELECT * FROM crm_pipelines WHERE business_id = ? AND default_marker = 1;

-- name: CreateDefaultPipeline :execresult
-- The unique (business_id, default_marker) key makes this a no-op when a
-- concurrent request created the default first.
INSERT IGNORE INTO crm_pipelines (business_id, name, position, default_marker) VALUES (?, ?, 0, 1);

-- name: CreatePipeline :execresult
INSERT INTO crm_pipelines (business_id, name, position) VALUES (?, ?, ?);

-- name: RenamePipeline :exec
UPDATE crm_pipelines SET name = ? WHERE id = ? AND business_id = ?;

-- name: DeletePipeline :execresult
DELETE FROM crm_pipelines WHERE id = ? AND business_id = ? AND default_marker IS NULL;

-- name: CountPipelines :one
SELECT COUNT(*) FROM crm_pipelines WHERE business_id = ?;

-- name: GetStage :one
SELECT * FROM crm_pipeline_stages WHERE id = ? AND business_id = ?;

-- name: CreateStage :execresult
INSERT INTO crm_pipeline_stages (business_id, pipeline_id, name, position, probability, kind)
VALUES (?, ?, ?, ?, ?, ?);

-- name: UpdateStage :exec
UPDATE crm_pipeline_stages SET name = ?, position = ?, probability = ?
WHERE id = ? AND business_id = ? AND pipeline_id = ?;

-- name: DeleteStage :execresult
DELETE FROM crm_pipeline_stages WHERE id = ? AND business_id = ? AND pipeline_id = ?;

-- name: DeletePipelineStages :exec
DELETE FROM crm_pipeline_stages WHERE business_id = ? AND pipeline_id = ?;
