CREATE TABLE IF NOT EXISTS hr_employee_documents (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    employee_id BIGINT NOT NULL,
    document_type VARCHAR(100) NOT NULL,
    title VARCHAR(255) NOT NULL,
    storage_mode VARCHAR(10) NOT NULL,
    file_url TEXT,
    external_url TEXT,
    expiry_date DATE DEFAULT NULL,
    uploaded_by BIGINT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_documents_owner_employee (owner_id, employee_id),
    INDEX idx_documents_owner_type (owner_id, document_type),
    INDEX idx_documents_owner_expiry (owner_id, expiry_date)
);
