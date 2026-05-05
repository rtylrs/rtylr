CREATE TABLE IF NOT EXISTS conversation_summaries (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    user_id INTEGER NOT NULL,
    session_id VARCHAR(255) NOT NULL,
    summary TEXT NOT NULL,
    key_topics JSON NULL,
    sentiment VARCHAR(50) NULL,
    importance_score FLOAT DEFAULT 1.0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_conversation_summaries_user_id ON conversation_summaries(user_id);
