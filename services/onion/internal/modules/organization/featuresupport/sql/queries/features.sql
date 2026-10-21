-- name: GetFeatureFlag :one
SELECT id, feature_key, description, enabled, rollout_percentage, variant_on, variant_off, created_at, updated_at
FROM feature_flags
WHERE feature_key = ?;

-- name: ListFeatureFlags :many
SELECT id, feature_key, description, enabled, rollout_percentage, variant_on, variant_off, created_at, updated_at
FROM feature_flags
ORDER BY feature_key ASC;

-- name: UpsertFeatureFlag :exec
INSERT INTO feature_flags (feature_key, description, enabled, rollout_percentage, variant_on, variant_off)
VALUES (?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
  description = VALUES(description),
  enabled = VALUES(enabled),
  rollout_percentage = VALUES(rollout_percentage),
  variant_on = VALUES(variant_on),
  variant_off = VALUES(variant_off);

-- name: DeleteFeatureFlag :exec
DELETE FROM feature_flags WHERE feature_key = ?;

-- name: ListFeatureTargetsByOwner :many
SELECT id, feature_key, owner_id, enabled, variant, created_at, updated_at
FROM feature_targets
WHERE owner_id = ?;

-- name: ListFeatureTargetsByKey :many
SELECT id, feature_key, owner_id, enabled, variant, created_at, updated_at
FROM feature_targets
WHERE feature_key = ?
ORDER BY owner_id ASC;

-- name: UpsertFeatureTarget :exec
INSERT INTO feature_targets (feature_key, owner_id, enabled, variant)
VALUES (?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
  enabled = VALUES(enabled),
  variant = VALUES(variant);

-- name: DeleteFeatureTarget :exec
DELETE FROM feature_targets WHERE feature_key = ? AND owner_id = ?;

-- name: ListAllRules :many
SELECT id, feature_key, priority, description, variant, rollout_percentage, enabled, created_at, updated_at
FROM feature_rules
WHERE enabled = 1
ORDER BY feature_key ASC, priority ASC, id ASC;

-- name: ListAllRuleConditions :many
SELECT id, rule_id, attribute, operator, value, created_at
FROM feature_rule_conditions;

-- name: ListRulesByKey :many
SELECT id, feature_key, priority, description, variant, rollout_percentage, enabled, created_at, updated_at
FROM feature_rules
WHERE feature_key = ?
ORDER BY priority ASC, id ASC;

-- name: CreateRule :execresult
INSERT INTO feature_rules (feature_key, priority, description, variant, rollout_percentage, enabled)
VALUES (?, ?, ?, ?, ?, ?);

-- name: UpdateRule :exec
UPDATE feature_rules
SET priority = ?, description = ?, variant = ?, rollout_percentage = ?, enabled = ?
WHERE id = ?;

-- name: DeleteRule :exec
DELETE FROM feature_rules WHERE id = ?;

-- name: DeleteConditionsByRule :exec
DELETE FROM feature_rule_conditions WHERE rule_id = ?;

-- cross-schema read from users — feature evaluation subject email
-- name: GetUserEmailByID :one
SELECT email FROM users WHERE id = ?;

-- name: ListConditionsByRuleIDs :many
SELECT id, rule_id, attribute, operator, value, created_at
FROM feature_rule_conditions
WHERE rule_id IN (sqlc.slice('rule_ids'))
ORDER BY rule_id ASC, id ASC;
