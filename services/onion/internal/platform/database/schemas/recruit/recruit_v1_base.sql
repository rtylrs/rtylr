-- Recruit product — tenant-scopes the existing `careers` table and
-- adds branding + submissions sidecars.
--
-- The legacy `careers` table was a single global list (the voxire team
-- page). That page is being retired and redirected to the recruit app,
-- so we extend the same table in place rather than duplicating the
-- schema. Existing rows get owner_id = 0 and is_published = false by
-- default; they won't surface on any tenant's careers page until a
-- business explicitly claims them.
--
-- Lives under its own directory (schemas/recruit/) instead of voxire/
-- so the sqlc config on the careers module doesn't try to parse these
-- statements. The ALTERs below aren't guarded with INFORMATION_SCHEMA
-- checks because sqlc can't parse prepared statements — migrate.sh
-- uses `mysql --force`, which continues past "duplicate column" errors
-- on re-runs. That's an acceptable trade: first run applies cleanly,
-- re-runs log a warning but still succeed.

-- ── careers: add tenant-scoping + publication columns ────────────────

ALTER TABLE careers ADD COLUMN owner_id BIGINT NOT NULL DEFAULT 0;
ALTER TABLE careers ADD COLUMN slug VARCHAR(120) NOT NULL DEFAULT '';
ALTER TABLE careers ADD COLUMN requirements TEXT NULL;
ALTER TABLE careers ADD COLUMN salary_min DECIMAL(12,2) NULL;
ALTER TABLE careers ADD COLUMN salary_max DECIMAL(12,2) NULL;
ALTER TABLE careers ADD COLUMN salary_currency VARCHAR(8) NULL;
ALTER TABLE careers ADD COLUMN is_published BOOLEAN NOT NULL DEFAULT false;
CREATE INDEX idx_careers_owner ON careers (owner_id);

-- Backfill slugs for any pre-existing rows. Without this, legacy rows
-- all default to slug='' and the unique index below would collide on
-- (owner_id=0, slug='').  Safe to re-run: the WHERE clause only touches
-- rows that are still empty.
UPDATE careers
  SET slug = LOWER(REPLACE(REPLACE(REPLACE(TRIM(title), ' ', '-'), '/', '-'), '.', ''))
  WHERE slug IS NULL OR slug = '';

CREATE UNIQUE INDEX uk_careers_owner_slug ON careers (owner_id, slug);

-- ── careers_customizations: per-business branding ────────────────────
--
-- One row per business that signs up for the recruit product. The slug
-- here is the path segment on the public careers site
-- (recruit.rtylr.com/<slug>); it must be unique globally.

CREATE TABLE IF NOT EXISTS careers_customizations (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
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
    business_id BIGINT NOT NULL DEFAULT 0,
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
