-- Available addons (system catalog)
CREATE TABLE IF NOT EXISTS addons (
    id VARCHAR(50) NOT NULL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    icon VARCHAR(50),
    price DECIMAL(10,2) DEFAULT 0.00,
    is_free BOOLEAN DEFAULT FALSE,
    is_default_enabled BOOLEAN DEFAULT FALSE,
    sort_order INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Addons enabled per business
CREATE TABLE IF NOT EXISTS business_addons (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    business_id BIGINT NOT NULL,
    addon_id VARCHAR(50) NOT NULL,
    is_enabled BOOLEAN DEFAULT TRUE,
    enabled_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    disabled_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY unique_business_addon (business_id, addon_id),
    INDEX idx_business_addons_business (business_id)
);

-- Seed the menu addon (disabled by default)
INSERT IGNORE INTO addons (id, name, description, icon, price, is_free, is_default_enabled, sort_order)
VALUES
    ('menu', 'Digital Menu', 'Customer-facing digital menu with QR codes, online ordering, and custom branding.', 'utensils-crossed', 0.00, true, false, 1),
    ('loyalty', 'Loyalty Program', 'Points-based loyalty and rewards to boost customer retention.', 'heart', 12.00, false, false, 2),
    ('ai_chat', 'AI Chat', 'AI-powered chat assistant for customer support and recommendations.', 'message-square', 0.00, true, false, 3),
    ('analytics', 'Advanced Analytics', 'Deep insights into sales, customer behavior, and menu performance.', 'bar-chart-2', 19.00, false, false, 4),
    ('custom_branding', 'Custom Branding', 'Fully customize your customer-facing experience with your brand.', 'paintbrush', 7.00, false, false, 5),
    ('multi_language', 'Multi-Language', 'Serve your menu and communications in multiple languages.', 'globe', 14.00, false, false, 6);
