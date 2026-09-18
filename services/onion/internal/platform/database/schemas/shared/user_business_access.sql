-- Business membership is part of the shared authorization schema.  Keep it in
-- the platform migration stream so tenant resolution is available before any
-- module integration test or handler can use it.
CREATE TABLE IF NOT EXISTS user_business_access (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL,
    role ENUM('owner','admin','viewer') NOT NULL DEFAULT 'owner',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uq_uba_user_business (user_id, business_id),
    INDEX idx_uba_user (user_id),
    INDEX idx_uba_business (business_id)
);
