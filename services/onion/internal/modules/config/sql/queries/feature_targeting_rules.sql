-- name: GetTargetingRulesByFeature :many
SELECT
    id,
    feature_flag_id,
    name,
    conditions,
    action,
    rollout_percentage,
    priority,
    enabled
FROM feature_targeting_rules
WHERE feature_flag_id = $1 AND enabled = true
ORDER BY priority DESC, id ASC;

-- name: GetTargetingRule :one
SELECT
    id,
    feature_flag_id,
    name,
    conditions,
    action,
    rollout_percentage,
    priority,
    enabled
FROM feature_targeting_rules
WHERE id = $1;
