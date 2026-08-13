CREATE TABLE IF NOT EXISTS crm_customer_timeline (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    customer_id BIGINT NOT NULL,
    type VARCHAR(50) NOT NULL,
    description TEXT,
    metadata JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    KEY idx_crm_customer_timeline_business (business_id),
    KEY idx_crm_customer_timeline_page (business_id, customer_id, created_at, id)
);
