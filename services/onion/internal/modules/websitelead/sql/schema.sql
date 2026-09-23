CREATE TABLE IF NOT EXISTS voxire_website_leads (
    id         BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    name       VARCHAR(255) NOT NULL,
    email      VARCHAR(255) NOT NULL,
    phone      VARCHAR(50),
    company    VARCHAR(255),
    service    VARCHAR(100),
    project    TEXT,
    budget     VARCHAR(50),
    source     VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

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
