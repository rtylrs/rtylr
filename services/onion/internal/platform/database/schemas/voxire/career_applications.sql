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
    created_at         TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_career_apps_job (job_id),
    INDEX idx_career_apps_status (status),
    INDEX idx_career_apps_job_status (job_id, status)
);

-- VoxHire admin accounts — self-contained login for the
-- hire.voxire.com/admin recruiter panel. Independent of the rtylr
-- team/user auth; admins are managed from within the panel itself.
CREATE TABLE IF NOT EXISTS voxhire_admins (
    id            BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    email         VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name          VARCHAR(255),
    created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_voxhire_admins_email (email)
);
