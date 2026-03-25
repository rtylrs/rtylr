CREATE TABLE IF NOT EXISTS user_memory (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    memory_type VARCHAR(100) NOT NULL,
    key VARCHAR(255) NOT NULL,
    value TEXT NOT NULL,
    confidence FLOAT DEFAULT 1.0,
    source_session VARCHAR(255),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id, memory_type, key)
);

CREATE INDEX IF NOT EXISTS idx_user_memory_user_id ON user_memory(user_id);
CREATE INDEX IF NOT EXISTS idx_user_memory_key ON user_memory(key);
