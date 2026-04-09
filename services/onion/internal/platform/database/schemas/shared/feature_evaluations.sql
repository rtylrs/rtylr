CREATE TABLE IF NOT EXISTS feature_evaluations (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NULL,
    feature_key VARCHAR(100) NOT NULL,
    enabled BOOLEAN NOT NULL,
    reason TEXT NULL,
    rule_id BIGINT NULL,
    user_context JSON NULL,
    evaluated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_feature_evaluations_rule_id
        FOREIGN KEY (rule_id) REFERENCES feature_targeting_rules(id)
);

CREATE INDEX idx_feature_evaluations_user_id ON feature_evaluations(user_id);
CREATE INDEX idx_feature_evaluations_feature_key ON feature_evaluations(feature_key);
CREATE INDEX idx_feature_evaluations_evaluated_at ON feature_evaluations(evaluated_at);
CREATE INDEX idx_feature_evaluations_enabled ON feature_evaluations(enabled);
