








-- Backfill slugs for any pre-existing rows. Without this, legacy rows
-- all default to slug='' and the unique index below would collide on
-- (owner_id=0, slug='').  Safe to re-run: the WHERE clause only touches
-- rows that are still empty.
UPDATE careers
  SET slug = LOWER(REPLACE(REPLACE(REPLACE(TRIM(title), ' ', '-'), '/', '-'), '.', ''))
  WHERE slug IS NULL OR slug = '';



CREATE TABLE IF NOT EXISTS `careers_customizations` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `owner_id` bigint NOT NULL,
  `business_id` bigint NOT NULL DEFAULT '0',
  `slug` varchar(80) COLLATE utf8mb4_unicode_ci NOT NULL,
  `brand_name` varchar(120) COLLATE utf8mb4_unicode_ci NOT NULL,
  `tagline` varchar(240) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `about_blurb` text COLLATE utf8mb4_unicode_ci,
  `logo_url` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `primary_color` varchar(16) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `template` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'minimal',
  `header_html` text COLLATE utf8mb4_unicode_ci,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `accent_color` varchar(16) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `bg_color` varchar(16) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `text_color` varchar(16) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `muted_text_color` varchar(16) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `font_pairing` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `hero_image_url` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `favicon_url` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `meta_title` varchar(240) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `meta_description` varchar(320) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `privacy_url` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `terms_url` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `ga_measurement_id` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `meta_pixel_id` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `require_consent` tinyint(1) NOT NULL DEFAULT '0',
  `notification_email` varchar(160) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `email_notifications` tinyint(1) NOT NULL DEFAULT '1',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_careers_customizations_owner` (`owner_id`),
  UNIQUE KEY `uk_careers_customizations_slug` (`slug`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `careers_submissions` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `career_id` bigint NOT NULL,
  `owner_id` bigint NOT NULL,
  `business_id` bigint NOT NULL DEFAULT '0',
  `candidate_name` varchar(160) COLLATE utf8mb4_unicode_ci NOT NULL,
  `email` varchar(160) COLLATE utf8mb4_unicode_ci NOT NULL,
  `phone` varchar(40) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `resume_url` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `cover_letter` text COLLATE utf8mb4_unicode_ci,
  `status` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'new',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `answers_json` json DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_careers_submissions_career` (`career_id`),
  KEY `idx_careers_submissions_owner` (`owner_id`),
  KEY `idx_careers_submissions_status` (`owner_id`,`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
