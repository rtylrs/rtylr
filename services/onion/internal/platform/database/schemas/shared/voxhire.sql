-- Voxhire lead-capture tables (company hiring leads + talent applications).

CREATE TABLE IF NOT EXISTS voxhire_company_leads (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    company_name VARCHAR(255) NOT NULL,
    contact_name VARCHAR(255) NOT NULL,
    contact_title VARCHAR(255) DEFAULT NULL,
    work_email VARCHAR(255) NOT NULL,
    phone VARCHAR(64) DEFAULT NULL,
    role VARCHAR(255) NOT NULL,
    headcount VARCHAR(32) DEFAULT NULL,
    timeline VARCHAR(32) DEFAULT NULL,
    service VARCHAR(32) DEFAULT NULL,
    message TEXT,
    agency_opt_in BOOLEAN NOT NULL DEFAULT FALSE,
    status VARCHAR(32) NOT NULL DEFAULT 'new',
    created_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_voxhire_company_leads_email (work_email),
    INDEX idx_voxhire_company_leads_status (status),
    INDEX idx_voxhire_company_leads_created (created_at)
);

CREATE TABLE IF NOT EXISTS voxhire_talent_applications (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    full_name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    phone VARCHAR(64) DEFAULT NULL,
    country_of_residence VARCHAR(128) DEFAULT NULL,
    nationality VARCHAR(128) DEFAULT NULL,
    track VARCHAR(64) DEFAULT NULL,
    years_experience VARCHAR(32) DEFAULT NULL,
    notice_period VARCHAR(32) DEFAULT NULL,
    open_to VARCHAR(255) DEFAULT NULL,
    linkedin_url VARCHAR(512) DEFAULT NULL,
    portfolio_url VARCHAR(512) DEFAULT NULL,
    pitch TEXT,
    consent BOOLEAN NOT NULL DEFAULT FALSE,
    cv_key VARCHAR(512) DEFAULT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'new',
    created_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_voxhire_talent_applications_email (email),
    INDEX idx_voxhire_talent_applications_track (track),
    INDEX idx_voxhire_talent_applications_status (status),
    INDEX idx_voxhire_talent_applications_created (created_at)
);
