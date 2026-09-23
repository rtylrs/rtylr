CREATE TABLE IF NOT EXISTS voxire_newsletter_subscribers (
    id            BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    email         VARCHAR(255) NOT NULL,
    source        VARCHAR(50)  NOT NULL,
    post_slug     VARCHAR(255),
    post_title    VARCHAR(255),
    post_category VARCHAR(100),
    language      VARCHAR(5)   NOT NULL DEFAULT 'en',
    created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_newsletter_email (email)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
