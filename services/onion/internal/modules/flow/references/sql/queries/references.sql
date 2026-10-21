-- name: ProjectExists :one
SELECT EXISTS(SELECT 1 FROM flow_projects WHERE business_id = ? AND id = ?) AS present;
