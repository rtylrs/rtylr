CREATE TABLE IF NOT EXISTS activity_log (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    app VARCHAR(20) NOT NULL DEFAULT 'hub' COMMENT 'Which app: hub, pos, erp, hr, crm, finance, flow',
    entity_type VARCHAR(50) NOT NULL,
    entity_id BIGINT,
    action VARCHAR(100) NOT NULL,
    description TEXT,
    user_id BIGINT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_activity_owner (owner_id),
    INDEX idx_activity_app (owner_id, app),
    INDEX idx_activity_created (owner_id, created_at)
);
