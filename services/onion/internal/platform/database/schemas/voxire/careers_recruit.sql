-- Recruit product — tenant-scopes the existing `careers` table and
-- adds branding + applications sidecars.
--
-- The legacy `careers` table was a single global list (the voxire team
-- page). That page is being retired and redirected to the recruit app,
-- so we extend the same table in place rather than duplicating the
-- schema. Existing rows get owner_id = 0 and is_published = false by
-- default; they won't surface on any tenant's careers page until a
-- business explicitly claims them.
--
-- MySQL 8 doesn't ship an IF NOT EXISTS clause for ADD COLUMN, so each
-- add is gated by an INFORMATION_SCHEMA check + prepared statement.
-- Verbose, but idempotent and re-runnable with no data loss.

-- ── careers: add tenant-scoping + publication columns ────────────────

SELECT COUNT(*) INTO @col_exists FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'careers' AND COLUMN_NAME = 'owner_id';
SET @sql := IF(@col_exists = 0, 'ALTER TABLE careers ADD COLUMN owner_id BIGINT NOT NULL DEFAULT 0', 'DO 0');
PREPARE st FROM @sql; EXECUTE st; DEALLOCATE PREPARE st;

SELECT COUNT(*) INTO @col_exists FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'careers' AND COLUMN_NAME = 'slug';
SET @sql := IF(@col_exists = 0, 'ALTER TABLE careers ADD COLUMN slug VARCHAR(120) NOT NULL DEFAULT ""', 'DO 0');
PREPARE st FROM @sql; EXECUTE st; DEALLOCATE PREPARE st;

SELECT COUNT(*) INTO @col_exists FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'careers' AND COLUMN_NAME = 'requirements';
SET @sql := IF(@col_exists = 0, 'ALTER TABLE careers ADD COLUMN requirements TEXT NULL', 'DO 0');
PREPARE st FROM @sql; EXECUTE st; DEALLOCATE PREPARE st;

SELECT COUNT(*) INTO @col_exists FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'careers' AND COLUMN_NAME = 'salary_min';
SET @sql := IF(@col_exists = 0, 'ALTER TABLE careers ADD COLUMN salary_min DECIMAL(12,2) NULL', 'DO 0');
PREPARE st FROM @sql; EXECUTE st; DEALLOCATE PREPARE st;

SELECT COUNT(*) INTO @col_exists FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'careers' AND COLUMN_NAME = 'salary_max';
SET @sql := IF(@col_exists = 0, 'ALTER TABLE careers ADD COLUMN salary_max DECIMAL(12,2) NULL', 'DO 0');
PREPARE st FROM @sql; EXECUTE st; DEALLOCATE PREPARE st;

SELECT COUNT(*) INTO @col_exists FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'careers' AND COLUMN_NAME = 'salary_currency';
SET @sql := IF(@col_exists = 0, 'ALTER TABLE careers ADD COLUMN salary_currency VARCHAR(8) NULL', 'DO 0');
PREPARE st FROM @sql; EXECUTE st; DEALLOCATE PREPARE st;

SELECT COUNT(*) INTO @col_exists FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'careers' AND COLUMN_NAME = 'is_published';
SET @sql := IF(@col_exists = 0, 'ALTER TABLE careers ADD COLUMN is_published BOOLEAN NOT NULL DEFAULT false', 'DO 0');
PREPARE st FROM @sql; EXECUTE st; DEALLOCATE PREPARE st;

-- Index owner_id for tenant-scoped queries. Idempotent via ANALYZE +
-- check on INFORMATION_SCHEMA.STATISTICS.
SELECT COUNT(*) INTO @idx_exists FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'careers' AND INDEX_NAME = 'idx_careers_owner';
SET @sql := IF(@idx_exists = 0, 'CREATE INDEX idx_careers_owner ON careers (owner_id)', 'DO 0');
PREPARE st FROM @sql; EXECUTE st; DEALLOCATE PREPARE st;

SELECT COUNT(*) INTO @idx_exists FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'careers' AND INDEX_NAME = 'uk_careers_owner_slug';
SET @sql := IF(@idx_exists = 0, 'CREATE UNIQUE INDEX uk_careers_owner_slug ON careers (owner_id, slug)', 'DO 0');
PREPARE st FROM @sql; EXECUTE st; DEALLOCATE PREPARE st;

-- ── careers_customizations: per-business branding ────────────────────
--
-- One row per business that signs up for the recruit product. The slug
-- here is the path segment on the public careers site
-- (recruit.rtylr.com/<slug>); it must be unique globally.

CREATE TABLE IF NOT EXISTS careers_customizations (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    slug VARCHAR(80) NOT NULL,
    brand_name VARCHAR(120) NOT NULL,
    tagline VARCHAR(240) NULL,
    about_blurb TEXT NULL,
    logo_url VARCHAR(500) NULL,
    primary_color VARCHAR(16) NULL,
    -- Template choice: "minimal" | "editorial" | "bold". The frontend
    -- resolves the component tree based on this value so each business
    -- gets a visually distinct careers page without us shipping bespoke
    -- code per customer.
    template VARCHAR(32) NOT NULL DEFAULT 'minimal',
    -- Optional custom HTML injected into the careers page header. Lets
    -- a business drop a navigation strip pointing back to their main
    -- site without waiting on us to build a full header editor.
    header_html TEXT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_careers_customizations_owner (owner_id),
    UNIQUE KEY uk_careers_customizations_slug (slug)
);

-- ── careers_submissions: candidate applications ──────────────────────
--
-- One row per application submitted through the public careers page.
-- `career_id` points at the opening being applied to; `owner_id` is
-- denormalised so tenant-scoped list queries don't need to join on
-- careers. resume_url is a plain URL for the MVP — the storage layer
-- is configurable per business (S3, R2, etc.) and the handler records
-- whatever URL the uploader returns.

CREATE TABLE IF NOT EXISTS careers_submissions (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    career_id BIGINT NOT NULL,
    owner_id BIGINT NOT NULL,
    candidate_name VARCHAR(160) NOT NULL,
    email VARCHAR(160) NOT NULL,
    phone VARCHAR(40) NULL,
    resume_url VARCHAR(500) NULL,
    cover_letter TEXT NULL,
    -- Status funnel: new → reviewed → interviewing → offered → hired
    -- (or → rejected at any stage). Kept as a VARCHAR so HR can rename
    -- stages later without a migration.
    status VARCHAR(32) NOT NULL DEFAULT 'new',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_careers_submissions_career (career_id),
    INDEX idx_careers_submissions_owner (owner_id),
    INDEX idx_careers_submissions_status (owner_id, status)
);
