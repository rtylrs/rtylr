CREATE TABLE IF NOT EXISTS conversation_messages (
    id SERIAL PRIMARY KEY,
    session_id VARCHAR(255) NOT NULL,
    user_id INTEGER NOT NULL,
    role VARCHAR(50) NOT NULL CHECK (role IN ('user', 'assistant', 'system')),
    content TEXT NOT NULL,
    tokens_used INTEGER,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_conversation_messages_session_id ON conversation_messages(session_id);
CREATE INDEX IF NOT EXISTS idx_conversation_messages_user_id ON conversation_messages(user_id);
CREATE INDEX IF NOT EXISTS idx_conversation_messages_created_at ON conversation_messages(created_at);
CREATE INDEX IF NOT EXISTS idx_conversation_messages_content_fts ON conversation_messages USING gin(to_tsvector('english', content));
