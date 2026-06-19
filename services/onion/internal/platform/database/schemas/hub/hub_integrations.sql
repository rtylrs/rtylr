CREATE TABLE IF NOT EXISTS hub_integrations (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    provider VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'connected',
    config JSON NOT NULL DEFAULT (JSON_OBJECT()),
    connected_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_hub_integrations_owner_id ON hub_integrations(owner_id);
