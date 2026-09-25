-- Candidate organization for the recruit ATS: reusable tags and
-- tenant-scoped talent pools with explicit membership rows.

CREATE TABLE IF NOT EXISTS careers_submission_tags (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    submission_id BIGINT NOT NULL,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    name VARCHAR(64) NOT NULL,
    normalized_name VARCHAR(64) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_submission_tags_name (business_id, submission_id, normalized_name),
    INDEX idx_submission_tags_submission (submission_id, created_at),
    INDEX idx_submission_tags_business_name (business_id, normalized_name)
);

CREATE TABLE IF NOT EXISTS careers_talent_pools (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    name VARCHAR(120) NOT NULL,
    description VARCHAR(500) NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_talent_pools_business_name (business_id, name),
    INDEX idx_talent_pools_business (business_id, created_at)
);

CREATE TABLE IF NOT EXISTS careers_talent_pool_members (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    pool_id BIGINT NOT NULL,
    submission_id BIGINT NOT NULL,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_talent_pool_membership (pool_id, submission_id),
    INDEX idx_talent_pool_members_business (business_id, pool_id),
    INDEX idx_talent_pool_members_submission (business_id, submission_id)
);
