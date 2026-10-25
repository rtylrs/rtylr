CREATE TABLE IF NOT EXISTS `addons` (
  `id` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `description` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
  `icon` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `price` decimal(10,2) DEFAULT '0.00',
  `is_free` tinyint(1) DEFAULT '0',
  `is_default_enabled` tinyint(1) DEFAULT '0',
  `sort_order` int DEFAULT '0',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `business_addons` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `business_id` bigint NOT NULL,
  `addon_id` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `is_enabled` tinyint(1) DEFAULT '1',
  `enabled_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `disabled_at` timestamp NULL DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_business_addon` (`business_id`,`addon_id`),
  KEY `idx_business_addons_business` (`business_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `manual_addon_grants` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `business_id` bigint NOT NULL,
  `addon_id` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `granted_by` bigint NOT NULL COMMENT 'console admin user_id who created the grant',
  `note` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
  `expires_at` timestamp NULL DEFAULT NULL COMMENT 'NULL = never expires',
  `revoked_at` timestamp NULL DEFAULT NULL,
  `enabled_by_grant` tinyint(1) NOT NULL DEFAULT '1' COMMENT 'true when the grant turned the addon on',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_manual_addon_grants_business` (`business_id`),
  KEY `idx_manual_addon_grants_lookup` (`business_id`,`addon_id`,`revoked_at`,`expires_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Seed the menu addon (disabled by default)
INSERT IGNORE INTO addons (id, name, description, icon, price, is_free, is_default_enabled, sort_order)
VALUES
    ('menu', 'Digital Menu', 'Customer-facing digital menu with QR codes, online ordering, and custom branding.', 'utensils-crossed', 0.00, true, false, 1),
    ('loyalty', 'Loyalty Program', 'Points-based loyalty and rewards to boost customer retention.', 'heart', 12.00, false, false, 2),
    ('analytics', 'Advanced Analytics', 'Deep insights into sales, customer behavior, and menu performance.', 'bar-chart-2', 19.00, false, false, 3),
    ('custom_branding', 'Custom Branding', 'Fully customize your customer-facing experience with your brand.', 'paintbrush', 7.00, false, false, 4),
    ('multi_language', 'Multi-Language', 'Serve your menu and communications in multiple languages.', 'globe', 14.00, false, false, 5),
    ('custom_domain', 'Custom Domain', 'Use your own domain for your customer-facing menu and ordering page.', 'globe', 9.00, false, false, 6),
    ('email_marketing', 'Email Marketing', 'Send campaigns, newsletters, and automated emails to your customers.', 'mail', 15.00, false, false, 7),
    ('sms', 'SMS Notifications', 'Send order updates and promotions via SMS to customers.', 'message-square', 12.00, false, false, 8),
    ('reservations', 'Reservations', 'Let customers book tables online with a reservation widget.', 'calendar', 0.00, true, false, 9),
    ('online_ordering', 'Online Ordering', 'Accept pickup and delivery orders from your website.', 'shopping-cart', 19.00, false, false, 10),
    ('kds', 'Kitchen Display', 'Real-time kitchen display system for order management.', 'monitor', 0.00, true, false, 11),
    ('api_access', 'API Access', 'Full REST API access for custom integrations and automation.', 'code', 29.00, false, false, 12),
    ('white_label', 'White Label', 'Remove rtylr branding from all customer-facing pages.', 'eye-off', 24.00, false, false, 13),
    ('priority_support', 'Priority Support', 'Dedicated support channel with guaranteed response times.', 'headphones', 19.00, false, false, 14),
    ('data_export', 'Data Export', 'Export all your business data in CSV, JSON, or Excel format.', 'download', 0.00, true, false, 15),
    ('insights', 'Insights', 'Cross-app analytics, revenue trends, and business intelligence dashboards.', 'bar-chart', 0.00, true, true, 16);
