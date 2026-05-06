CREATE TABLE IF NOT EXISTS cdn_files (
    id                BIGINT       NOT NULL AUTO_INCREMENT PRIMARY KEY,
    file_uuid         VARCHAR(36)  NOT NULL,
    storage_key       VARCHAR(512) NOT NULL,
    mime_type         VARCHAR(128) NOT NULL,
    size              BIGINT       NOT NULL,
    original_filename VARCHAR(512) NOT NULL,
    created_at        TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uq_cdn_files_uuid (file_uuid),
    UNIQUE KEY uq_cdn_files_storage_key (storage_key)
);

CREATE INDEX idx_cdn_files_created_at ON cdn_files(created_at);
