CREATE TABLE IF NOT EXISTS feature_targeting_rules (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    feature_flag_id INTEGER NOT NULL,
    name VARCHAR(200) NOT NULL,
    conditions JSON NOT NULL,
    action VARCHAR(20) NOT NULL CHECK (action IN ('enable', 'disable', 'rollout')),
    rollout_percentage INTEGER DEFAULT 0 CHECK (rollout_percentage >= 0 AND rollout_percentage <= 100),
    priority INTEGER DEFAULT 0,
    enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by INT NULL,
    updated_by INT NULL,
    CONSTRAINT fk_feature_targeting_rules_feature_flag_id
        FOREIGN KEY (feature_flag_id) REFERENCES feature_flags(id) ON DELETE CASCADE
);

CREATE INDEX idx_feature_targeting_rules_feature_flag_id ON feature_targeting_rules(feature_flag_id);
CREATE INDEX idx_feature_targeting_rules_enabled ON feature_targeting_rules(enabled);
CREATE INDEX idx_feature_targeting_rules_priority ON feature_targeting_rules(priority DESC);
