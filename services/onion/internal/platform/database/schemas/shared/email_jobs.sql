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

    INDEX idx_pending_priority (status, priority DESC, scheduled_at ASC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
