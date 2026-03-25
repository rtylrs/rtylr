CREATE TABLE user_feature_overrides (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    feature_key VARCHAR(100) NOT NULL,
    enabled BOOLEAN NOT NULL,
    reason TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    created_by INT,
    UNIQUE(user_id, feature_key)
);

CREATE INDEX idx_user_feature_overrides_user_id ON user_feature_overrides(user_id);
CREATE INDEX idx_user_feature_overrides_feature_key ON user_feature_overrides(feature_key);
CREATE INDEX idx_user_feature_overrides_enabled ON user_feature_overrides(enabled);