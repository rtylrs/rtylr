-- Local schema for sqlc (production DDL from voxhire auto-migrate).
CREATE TABLE IF NOT EXISTS voxhire_company_leads (
    id            BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    company_name  VARCHAR(255) NOT NULL,
    contact_name  VARCHAR(255) NOT NULL,
    contact_title VARCHAR(255),
    work_email    VARCHAR(255) NOT NULL,
    phone         VARCHAR(64),
    role          VARCHAR(255) NOT NULL,
    headcount     VARCHAR(32),
    timeline      VARCHAR(32),
    service       VARCHAR(32),
    message       TEXT,
    agency_opt_in TINYINT(1) NOT NULL DEFAULT 0,
    status        VARCHAR(32) NOT NULL DEFAULT 'new',
    created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS voxhire_talent_applications (
    id                   BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    full_name            VARCHAR(255) NOT NULL,
    email                VARCHAR(255) NOT NULL,
    phone                VARCHAR(64),
    country_of_residence VARCHAR(128),
    nationality          VARCHAR(128),
    track                VARCHAR(64),
    years_experience     VARCHAR(32),
    notice_period        VARCHAR(32),
    open_to              VARCHAR(255),
    `current_role`       VARCHAR(255),
    desired_roles        TEXT,
    salary_expectation   VARCHAR(128),
    work_authorization   VARCHAR(255),
    work_history         TEXT,
    education            TEXT,
    linkedin_url         VARCHAR(512),
    portfolio_url        VARCHAR(512),
    pitch                TEXT,
    consent              TINYINT(1) NOT NULL DEFAULT 0,
    cv_key               VARCHAR(512),
    cv_processing_status VARCHAR(32) NOT NULL DEFAULT 'pending',
    status               VARCHAR(32) NOT NULL DEFAULT 'new',
    created_at           TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE IF NOT EXISTS voxhire_admins (
    id            BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    email         VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name          VARCHAR(255),
    created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_voxhire_admins_email (email)
);

CREATE TABLE IF NOT EXISTS voxhire_application_events (
    id             BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    application_id BIGINT NOT NULL,
    admin_id       BIGINT,
    kind           VARCHAR(64) NOT NULL,
    title          VARCHAR(255) NOT NULL,
    detail         TEXT,
    created_at     TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_voxhire_events_application (application_id)
);

CREATE TABLE IF NOT EXISTS voxhire_follow_up_tasks (
    id             BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    application_id BIGINT NOT NULL,
    title          VARCHAR(255) NOT NULL,
    status         VARCHAR(32) NOT NULL DEFAULT 'open',
    due_at         TIMESTAMP NULL,
    completed_at   TIMESTAMP NULL,
    created_at     TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_voxhire_task_application_title (application_id, title)
);

CREATE TABLE IF NOT EXISTS voxhire_shortlists (
    id         BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    job_id     BIGINT NOT NULL,
    name       VARCHAR(255) NOT NULL,
    status     VARCHAR(32) NOT NULL DEFAULT 'draft',
    notes      TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS voxhire_shortlist_items (
    id             BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    shortlist_id   BIGINT NOT NULL,
    application_id BIGINT NOT NULL,
    sort_order     INT NOT NULL DEFAULT 0,
    notes          TEXT,
    created_at     TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_voxhire_shortlist_application (shortlist_id, application_id)
);

CREATE TABLE IF NOT EXISTS careers (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    business_id BIGINT NOT NULL DEFAULT 0,
    owner_id BIGINT NOT NULL DEFAULT 0,
    slug VARCHAR(120) NOT NULL DEFAULT '',
    title VARCHAR(255) NOT NULL,
    department VARCHAR(100) NULL,
    location VARCHAR(100) NULL,
    employment_type VARCHAR(50) NULL,
    work_mode VARCHAR(20) NULL,
    seniority VARCHAR(50) NULL,
    details TEXT NOT NULL,
    requirements TEXT NULL,
    salary_min DECIMAL(12,2) NULL,
    salary_max DECIMAL(12,2) NULL,
    salary_currency VARCHAR(8) NULL,
    is_published BOOLEAN NOT NULL DEFAULT false,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS voxire_career_applications (
    id                 BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    job_id             BIGINT NOT NULL,
    name               VARCHAR(255) NOT NULL,
    email              VARCHAR(255) NOT NULL,
    phone              VARCHAR(100),
    work_history       TEXT,
    education          TEXT,
    experience_summary TEXT,
    cover_letter       TEXT,
    cv_key             VARCHAR(500),
    location           VARCHAR(160),
    salary_expectation DECIMAL(12,2),
    salary_currency    VARCHAR(8) DEFAULT 'USD',
    status             VARCHAR(32) NOT NULL DEFAULT 'new',
    cv_processing_status VARCHAR(32) NOT NULL DEFAULT 'pending',
    cv_processing_error TEXT,
    cv_extracted_text MEDIUMTEXT,
    cv_summary TEXT,
    cv_profile_json TEXT,
    cv_processed_at TIMESTAMP NULL,
    review_rating TINYINT NOT NULL DEFAULT 0,
    review_notes TEXT,
    next_step VARCHAR(255),
    reviewed_at TIMESTAMP NULL,
    created_at         TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
