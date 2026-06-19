-- name: GetPerformanceConfig :one
SELECT id, owner_id, business_id, rating_scale_type, rating_min, rating_max, COALESCE(rating_labels, '[]') AS rating_labels, peer_feedback_enabled, created_at, updated_at
FROM hr_performance_config
WHERE business_id = ?;

-- name: UpsertPerformanceConfig :execresult
INSERT INTO hr_performance_config (owner_id, business_id, rating_scale_type, rating_min, rating_max, rating_labels, peer_feedback_enabled)
VALUES (?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
    rating_scale_type = VALUES(rating_scale_type),
    rating_min = VALUES(rating_min),
    rating_max = VALUES(rating_max),
    rating_labels = VALUES(rating_labels),
    peer_feedback_enabled = VALUES(peer_feedback_enabled);
