CREATE TABLE IF NOT EXISTS user_app_permissions (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    business_id BIGINT NULL,
    owner_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    apps JSON NOT NULL COMMENT 'Array of app IDs: pos, erp, hr, crm, finance, flow',
    role ENUM('owner','admin','operator','billing','menu_manager','store_manager','staff','viewer') DEFAULT 'staff',
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY unique_user_perm (owner_id, user_id),
    INDEX idx_perms_business (business_id)
);
