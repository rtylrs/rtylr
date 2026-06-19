CREATE TABLE IF NOT EXISTS hub_notifications (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    user_id BIGINT NOT NULL,
    title VARCHAR(255) NOT NULL,
    body TEXT NULL,
    type VARCHAR(50) NOT NULL DEFAULT 'info',
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    link VARCHAR(2048) NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_hub_notifications_owner_id ON hub_notifications(owner_id);
CREATE INDEX idx_hub_notifications_user_id ON hub_notifications(user_id);
CREATE INDEX idx_hub_notifications_is_read ON hub_notifications(is_read);
