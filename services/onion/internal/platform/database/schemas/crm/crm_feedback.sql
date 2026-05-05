CREATE TABLE IF NOT EXISTS crm_feedback (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    customer_id BIGINT NOT NULL,
    type ENUM('nps','review') NOT NULL,
    score INT DEFAULT 0,
    comment TEXT,
    source VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
