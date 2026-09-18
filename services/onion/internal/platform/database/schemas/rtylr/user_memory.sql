CREATE TABLE IF NOT EXISTS user_memory (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    user_id INTEGER NOT NULL,
    memory_type VARCHAR(100) NOT NULL,
    `key` VARCHAR(255) NOT NULL,
    value TEXT NOT NULL,
    confidence FLOAT DEFAULT 1.0,
    source_session VARCHAR(255) NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, memory_type, `key`)
);

CREATE INDEX idx_user_memory_user_id ON user_memory(user_id);
CREATE INDEX idx_user_memory_key ON user_memory(`key`);
-- idx_user_memory_search is added by zz_user_memory_search_indexes.sql; adding
-- it here too fails a fresh migration run with a duplicate-key error.
