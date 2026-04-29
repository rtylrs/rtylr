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
    created_at         TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
