-- name: LogFeatureEvaluation :exec
INSERT INTO feature_evaluations (
    user_id,
    feature_key,
    enabled,
    reason,
    rule_id,
    user_context
) VALUES (
    ?, ?, ?, ?, ?, ?
);
