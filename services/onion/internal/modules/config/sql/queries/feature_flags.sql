-- name: GetFeatureFlag :one
SELECT
    id,
    `key`,
    name,
    description,
    enabled,
    rollout_percentage
FROM feature_flags
WHERE `key` = ?;

-- name: GetAllFeatureFlags :many
SELECT
    id,
    `key`,
    name,
    description,
    enabled,
    rollout_percentage
FROM feature_flags
ORDER BY name;

-- name: GetEnabledFeatureFlags :many
SELECT
    id,
    `key`,
    name,
    description,
    enabled,
    rollout_percentage
FROM feature_flags
WHERE enabled = true
ORDER BY name;

-- name: GetFeatureFlagsWithRules :many
SELECT
    ff.id AS feature_flag_id,
    ff.`key` AS feature_key,
    ff.name AS feature_name,
    ff.description AS feature_description,
    ff.enabled AS feature_enabled,
    ff.rollout_percentage AS feature_rollout_percentage,
    ftr.id AS rule_id,
    ftr.name AS rule_name,
    ftr.conditions AS rule_conditions,
    ftr.action AS rule_action,
    ftr.rollout_percentage AS rule_rollout_percentage,
    ftr.priority AS rule_priority,
    ftr.enabled AS rule_enabled
FROM feature_flags ff
LEFT JOIN feature_targeting_rules ftr ON ff.id = ftr.feature_flag_id AND ftr.enabled = true
WHERE ff.enabled = true
ORDER BY ff.name, ftr.priority DESC, ftr.id ASC;
