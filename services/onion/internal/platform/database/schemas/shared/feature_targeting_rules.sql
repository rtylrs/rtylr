CREATE TABLE feature_targeting_rules (
    id SERIAL PRIMARY KEY,
    feature_flag_id INTEGER NOT NULL REFERENCES feature_flags(id) ON DELETE CASCADE,
    name VARCHAR(200) NOT NULL,
    conditions JSONB NOT NULL,
    action VARCHAR(20) NOT NULL CHECK (action IN ('enable', 'disable', 'rollout')),
    rollout_percentage INTEGER DEFAULT 0 CHECK (rollout_percentage >= 0 AND rollout_percentage <= 100),
    priority INTEGER DEFAULT 0,
    enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    created_by INT,
    updated_by INT
);

CREATE INDEX idx_feature_targeting_rules_feature_flag_id ON feature_targeting_rules(feature_flag_id);
CREATE INDEX idx_feature_targeting_rules_enabled ON feature_targeting_rules(enabled);
CREATE INDEX idx_feature_targeting_rules_priority ON feature_targeting_rules(priority DESC);
CREATE INDEX idx_feature_targeting_rules_conditions ON feature_targeting_rules USING GIN (conditions);