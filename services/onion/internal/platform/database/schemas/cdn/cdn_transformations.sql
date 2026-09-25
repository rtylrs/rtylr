CREATE TABLE IF NOT EXISTS cdn_transformations (
    id          BIGINT       NOT NULL AUTO_INCREMENT PRIMARY KEY,
    file_id     BIGINT       NOT NULL,
    cache_key   VARCHAR(64)  NOT NULL,
    width       INT          NULL,
    height      INT          NULL,
    format      VARCHAR(16)  NULL,
    quality     INT          NULL,
    fit         VARCHAR(16)  NULL,
    storage_key VARCHAR(512) NOT NULL,
    created_at  TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uq_cdn_transformations_cache_key (cache_key),
    CONSTRAINT fk_cdn_transformations_file
        FOREIGN KEY (file_id) REFERENCES cdn_files(id) ON DELETE CASCADE
);

CREATE INDEX idx_cdn_transformations_file_id   ON cdn_transformations(file_id);
CREATE INDEX idx_cdn_transformations_cache_key ON cdn_transformations(cache_key);
