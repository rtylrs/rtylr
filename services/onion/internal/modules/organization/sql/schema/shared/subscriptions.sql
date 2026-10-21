CREATE TABLE IF NOT EXISTS `plans` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `slug` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `apps` json NOT NULL COMMENT 'Array of app IDs included in plan',
  `price_per_seat_cents` int NOT NULL DEFAULT '0',
  `paddle_price_id` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `trial_days` int NOT NULL DEFAULT '14',
  `tagline` varchar(200) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `annual_price_per_seat_cents` int NOT NULL DEFAULT '0',
  `max_seats` int DEFAULT NULL,
  `max_locations` int DEFAULT NULL,
  `ai_credits_per_month` int NOT NULL DEFAULT '0',
  `features` json DEFAULT NULL,
  `tier_order` int NOT NULL DEFAULT '100',
  `is_solo` tinyint(1) NOT NULL DEFAULT '0',
  `is_featured` tinyint(1) NOT NULL DEFAULT '0',
  `is_active` tinyint(1) DEFAULT '1',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `slug` (`slug`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `subscriptions` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `owner_id` bigint NOT NULL,
  `plan_id` bigint NOT NULL,
  `status` enum('active','trial','past_due','cancelled','expired') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT 'trial',
  `seat_count` int NOT NULL DEFAULT '1',
  `price_per_seat_cents` int NOT NULL DEFAULT '0',
  `annual_price_per_seat_cents` int NOT NULL DEFAULT '0',
  `trial_ends_at` timestamp NULL DEFAULT NULL,
  `next_billing_date` date DEFAULT NULL,
  `current_period_ends_at` timestamp NULL DEFAULT NULL,
  `paddle_subscription_id` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `paddle_customer_id` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `owner_id` (`owner_id`),
  KEY `idx_sub_owner` (`owner_id`),
  KEY `idx_sub_plan` (`plan_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `seat_periods` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `subscription_id` bigint NOT NULL,
  `owner_id` bigint NOT NULL,
  `seat_count` int NOT NULL,
  `price_per_seat_cents` int NOT NULL,
  `period_start` date NOT NULL,
  `period_end` date DEFAULT NULL COMMENT 'NULL = currently active period',
  `billing_month` date NOT NULL COMMENT 'First day of the month this period belongs to',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_sp_subscription` (`subscription_id`),
  KEY `idx_sp_owner_month` (`owner_id`,`billing_month`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ── Retire old tier names ─────────────────────────────────────────────────
UPDATE plans SET is_active = FALSE WHERE slug IN ('starter', 'enterprise');

-- ── Seed new tier + solo plans ────────────────────────────────────────────
-- Prices are in cents. Annual prices reflect the 20%-off annual pricing.
--   Free:   $0 / $0       · 10 seats · 1 location  · all 7 apps · 1,000 AI credits
--   Pro:    $6.99 / $4.99 · 30 seats · 5 locations · all 7 apps · 10,000 + 500/seat
--   Scale:  $4.99 / $2.99 · unlimited seats/locations · 50,000 + 1,000/seat
-- Solo plans: single-app starter tiers for individuals.

INSERT INTO plans
  (name, slug, apps, price_per_seat_cents, annual_price_per_seat_cents, trial_days,
   tagline, max_seats, max_locations, ai_credits_per_month, tier_order, is_solo, is_featured, features)
VALUES
  ('Free', 'free',
    '["dash","pos","erp","hr","crm","finance","flow","insights"]',
    0, 0, 0,
    'Everything you need to run a single location.',
    10, 1, 1000, 10, FALSE, FALSE,
    JSON_ARRAY('Up to 10 team members', 'All 7 apps included', '1 location', '0% transaction fees under $10k/mo', '1,000 AI credits / month')),
  ('Pro', 'pro',
    '["dash","pos","erp","hr","crm","finance","flow","insights"]',
    699, 499, 14,
    'Growing teams scaling across multiple locations.',
    30, 5, 10000, 20, FALSE, TRUE,
    JSON_ARRAY('Up to 30 team members', 'Up to 5 locations', '0% transaction fees under $100k/mo', '10,000 AI credits / month + 500 per seat', 'Priority email + chat support')),
  ('Scale', 'scale',
    '["dash","pos","erp","hr","crm","finance","flow","insights"]',
    499, 299, 14,
    'Unlimited seats and every feature unlocked.',
    NULL, NULL, 50000, 30, FALSE, FALSE,
    JSON_ARRAY('Unlimited team members', 'Unlimited locations', '0% transaction fees (no cap)', '50,000 AI credits / month + 1,000 per seat', 'SSO, audit logs, dedicated success manager')),
  ('Flow Solo', 'flow_solo', '["flow"]', 0, 0, 14,
    'Kanban, tasks, and automations for solo operators.',
    5, 1, 500, 40, TRUE, FALSE,
    JSON_ARRAY('Up to 5 team members', 'Unlimited boards + tasks', 'Daily AI credits', 'Free forever for individuals')),
  ('Finance Solo', 'finance_solo', '["finance"]', 0, 0, 14,
    'Invoices, expenses, and a real P&L for freelancers.',
    2, 1, 500, 41, TRUE, FALSE,
    JSON_ARRAY('Up to 2 team members', 'Unlimited invoices + expenses', 'Live P&L', 'Free forever for individuals')),
  ('POS Solo', 'pos_solo', '["pos"]', 0, 0, 14,
    'A register for single-location restaurants and retail.',
    3, 1, 500, 42, TRUE, FALSE,
    JSON_ARRAY('Up to 3 team members', '1 location', 'Offline mode included', '0% transaction fees under $5k/mo')),
  ('ERP Solo', 'erp_solo', '["erp"]', 0, 0, 14,
    'Inventory and purchasing for operators tracking SKUs only.',
    3, 1, 500, 43, TRUE, FALSE,
    JSON_ARRAY('Up to 3 team members', 'Unlimited SKUs', 'Supplier + PO tracking', 'Multi-unit variants')),
  ('HR Solo', 'hr_solo', '["hr"]', 0, 0, 14,
    'Scheduling and payroll for small teams.',
    10, 1, 500, 44, TRUE, FALSE,
    JSON_ARRAY('Up to 10 team members', 'Scheduling + time-off', 'Payroll included', 'Compliance flags')),
  ('CRM Solo', 'crm_solo', '["crm"]', 0, 0, 14,
    'Customer profiles and campaigns for owner-operators.',
    3, 1, 500, 45, TRUE, FALSE,
    JSON_ARRAY('Up to 3 team members', 'Unlimited customer profiles', 'Segment builder + campaigns', 'Loyalty points engine'))
ON DUPLICATE KEY UPDATE
  name = VALUES(name),
  apps = VALUES(apps),
  price_per_seat_cents = VALUES(price_per_seat_cents),
  annual_price_per_seat_cents = VALUES(annual_price_per_seat_cents),
  trial_days = VALUES(trial_days),
  tagline = VALUES(tagline),
  max_seats = VALUES(max_seats),
  max_locations = VALUES(max_locations),
  ai_credits_per_month = VALUES(ai_credits_per_month),
  tier_order = VALUES(tier_order),
  is_solo = VALUES(is_solo),
  is_featured = VALUES(is_featured),
  features = VALUES(features),
  is_active = TRUE;
