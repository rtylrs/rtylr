CREATE TABLE IF NOT EXISTS conversation_summaries (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    session_id VARCHAR(255) NOT NULL,
    summary TEXT NOT NULL,
    key_topics TEXT[],
    sentiment VARCHAR(50),
    importance_score FLOAT DEFAULT 1.0,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_conversation_summaries_user_id ON conversation_summaries(user_id);
