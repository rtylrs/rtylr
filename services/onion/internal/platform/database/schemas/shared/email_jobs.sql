CREATE TABLE IF NOT EXISTS email_jobs (
    id           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    priority     TINYINT        NOT NULL DEFAULT 1,
    status       ENUM('pending','processing','sent','failed') NOT NULL DEFAULT 'pending',
    to_email     VARCHAR(320)   NOT NULL,
    subject      VARCHAR(998)   NOT NULL,
    html         MEDIUMTEXT     NOT NULL,
    text_body    MEDIUMTEXT,
    from_email   VARCHAR(320),
    reply_to     VARCHAR(320),
    attempts     TINYINT UNSIGNED NOT NULL DEFAULT 0,
    max_attempts TINYINT UNSIGNED NOT NULL DEFAULT 3,
    last_error   TEXT,
    scheduled_at TIMESTAMP      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    sent_at      TIMESTAMP      NULL,
    created_at   TIMESTAMP      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMP      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    job_id       VARCHAR(36)    NULL DEFAULT NULL,
    -- Opt-in idempotency for a caller that can name the message. NULL means
    -- no deduplication, which is every caller that does not set one, and NULL
    -- does not collide in a unique index. See zz_email_jobs_dedupe.sql, which
    -- adds the same column and index to databases created before it.
    dedupe_key   VARCHAR(191)   NULL DEFAULT NULL,

    INDEX idx_pending_priority (status, priority DESC, scheduled_at ASC),
    INDEX idx_job_id (job_id),
    UNIQUE KEY uk_email_jobs_dedupe (dedupe_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
