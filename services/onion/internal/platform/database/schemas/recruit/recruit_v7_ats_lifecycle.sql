CREATE TABLE IF NOT EXISTS careers_submission_interviews (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    submission_id BIGINT NOT NULL,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    interview_type VARCHAR(32) NOT NULL DEFAULT 'video',
    title VARCHAR(255) NOT NULL,
    scheduled_at DATETIME NOT NULL,
    duration_minutes INT NOT NULL DEFAULT 60,
    timezone VARCHAR(64) NOT NULL DEFAULT 'UTC',
    meeting_url VARCHAR(500) NULL,
    interviewer_name VARCHAR(160) NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'scheduled',
    feedback TEXT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_submission_interviews_submission (submission_id, scheduled_at),
    INDEX idx_submission_interviews_business_time (business_id, scheduled_at)
);

CREATE TABLE IF NOT EXISTS careers_submission_notes (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    submission_id BIGINT NOT NULL,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    author_id BIGINT NOT NULL DEFAULT 0,
    body TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_submission_notes_submission (submission_id, created_at),
    INDEX idx_submission_notes_business (business_id, created_at)
);

CREATE TABLE IF NOT EXISTS careers_submission_offers (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    submission_id BIGINT NOT NULL,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    title VARCHAR(255) NOT NULL,
    salary DECIMAL(12,2) NULL,
    salary_currency VARCHAR(8) NOT NULL DEFAULT 'USD',
    start_date DATE NULL,
    expires_at DATE NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'draft',
    terms TEXT NULL,
    sent_at DATETIME NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_submission_offers_submission (submission_id, created_at),
    INDEX idx_submission_offers_business_status (business_id, status)
);
