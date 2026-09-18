CREATE TABLE IF NOT EXISTS voxire_website_leads (
    id          BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    email       VARCHAR(255) NOT NULL,
    phone       VARCHAR(50),
    company     VARCHAR(255),
    service     VARCHAR(100),
    project     TEXT,
    budget      VARCHAR(50),
    source      VARCHAR(100),
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
