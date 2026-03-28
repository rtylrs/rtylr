-- name: GetUserFeatureOverrides :many
SELECT
    id,
    user_id,
    feature_key,
    enabled,
    reason
FROM user_feature_overrides
WHERE user_id = ?
ORDER BY feature_key;

-- name: GetUserFeatureOverride :one
SELECT
    id,
    user_id,
    feature_key,
    enabled,
    reason
FROM user_feature_overrides
WHERE user_id = ? AND feature_key = ?;
