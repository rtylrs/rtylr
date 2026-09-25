CREATE TABLE IF NOT EXISTS crm_customer_timeline (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    customer_id BIGINT NOT NULL,
    type VARCHAR(50) NOT NULL,
    description TEXT,
    metadata JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    KEY idx_crm_customer_timeline_business (business_id)
    -- idx_crm_customer_timeline_page is added by
    -- zz_crm_customer_timeline_page_index.sql; declaring it here as well fails
    -- a fresh migration run with a duplicate-key error.
);
