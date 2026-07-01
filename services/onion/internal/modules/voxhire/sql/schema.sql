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
