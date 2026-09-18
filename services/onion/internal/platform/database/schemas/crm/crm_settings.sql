CREATE TABLE IF NOT EXISTS crm_settings (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    default_consent BOOLEAN DEFAULT TRUE,
    from_email VARCHAR(255),
    from_name VARCHAR(255),
    sms_sender_id VARCHAR(50),
    timezone VARCHAR(100) DEFAULT 'UTC',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uq_crm_settings_business (business_id)
);
