CREATE TABLE feature_evaluations (
    id SERIAL PRIMARY KEY,
    user_id INTEGER,
    feature_key VARCHAR(100) NOT NULL,
    enabled BOOLEAN NOT NULL,
    reason TEXT,
    rule_id INTEGER REFERENCES feature_targeting_rules(id),
    user_context JSONB,
    evaluated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_feature_evaluations_user_id ON feature_evaluations(user_id);
CREATE INDEX idx_feature_evaluations_feature_key ON feature_evaluations(feature_key);
CREATE INDEX idx_feature_evaluations_evaluated_at ON feature_evaluations(evaluated_at);
CREATE INDEX idx_feature_evaluations_enabled ON feature_evaluations(enabled);

-- Partition by month for better performance on large datasets
-- CREATE TABLE feature_evaluations_y2024m01 PARTITION OF feature_evaluations
-- FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');