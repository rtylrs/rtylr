-- name: GetUserFeatureOverrides :many
SELECT
    id,
    user_id,
    feature_key,
    enabled,
    reason
FROM user_feature_overrides
WHERE user_id = $1
ORDER BY feature_key;

-- name: GetUserFeatureOverride :one
SELECT
    id,
    user_id,
    feature_key,
    enabled,
    reason
FROM user_feature_overrides
WHERE user_id = $1 AND feature_key = $2;

