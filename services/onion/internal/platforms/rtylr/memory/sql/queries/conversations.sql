-- name: CreateChatSession :one
INSERT INTO chat_sessions (user_id, session_id, title, metadata)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetChatSession :one
SELECT * FROM chat_sessions 
WHERE session_id = $1 AND user_id = $2;

-- name: UpdateChatSession :one
UPDATE chat_sessions 
SET title = $3, summary = $4, metadata = $5, last_activity = NOW()
WHERE session_id = $1 AND user_id = $2
RETURNING *;

-- name: ListUserSessions :many
SELECT * FROM chat_sessions 
WHERE user_id = $1 
ORDER BY last_activity DESC 
LIMIT $2 OFFSET $3;

-- name: CreateMessage :one
INSERT INTO conversation_messages (session_id, user_id, role, content, tokens_used, metadata)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetSessionMessages :many
SELECT * FROM conversation_messages 
WHERE session_id = $1 
ORDER BY created_at ASC;

-- name: GetRecentMessages :many
SELECT * FROM conversation_messages 
WHERE user_id = $1 
ORDER BY created_at DESC 
LIMIT $2;

-- name: GetMessagesByTopic :many
SELECT cm.*, cs.title AS session_title
FROM conversation_messages cm
JOIN chat_sessions cs ON cm.session_id = cs.session_id
WHERE cm.user_id = $1 
AND to_tsvector('english', cm.content) @@ plainto_tsquery('english', $2)
ORDER BY cm.created_at DESC
LIMIT $3;

-- name: CreateConversationSummary :one
INSERT INTO conversation_summaries (user_id, session_id, summary, key_topics, sentiment, importance_score)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetUserSummaries :many
SELECT * FROM conversation_summaries 
WHERE user_id = $1 
ORDER BY importance_score DESC, created_at DESC;

-- name: StoreUserMemory :one
INSERT INTO user_memory (user_id, memory_type, key, value, confidence, source_session)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (user_id, memory_type, key) 
DO UPDATE SET 
    value = EXCLUDED.value,
    confidence = EXCLUDED.confidence,
    updated_at = NOW()
RETURNING *;

-- name: GetUserMemory :many
SELECT * FROM user_memory 
WHERE user_id = $1 
AND (memory_type = $2 OR $2 = '')
ORDER BY confidence DESC, updated_at DESC;

-- name: SearchUserMemory :many
SELECT * FROM user_memory 
WHERE user_id = $1 
AND (
    key ILIKE $2 OR 
    value ILIKE $2 OR
    memory_type ILIKE $2
)
ORDER BY confidence DESC;

-- name: GetContextForSession :many
WITH recent_messages AS (
    SELECT * FROM conversation_messages 
    WHERE conversation_messages.session_id = $1 
    ORDER BY created_at DESC 
    LIMIT 20
),
session_summary AS (
    SELECT summary, key_topics FROM conversation_summaries 
    WHERE conversation_summaries.session_id = $1 
    ORDER BY created_at DESC 
    LIMIT 1
),
user_memories AS (
    SELECT memory_type, key, value FROM user_memory 
    WHERE user_memory.user_id = $2 
    AND confidence > 0.7
    ORDER BY confidence DESC 
    LIMIT 10
)
SELECT 
    type,
    COALESCE(role, '') AS role,
    content,
    created_at,
    summary,
    key_topics,
    memory_type,
    memory_key,
    memory_value
FROM (
    SELECT 
        'message' AS type,
        recent_messages.role,
        recent_messages.content,
        recent_messages.created_at,
        NULL AS summary,
        NULL AS key_topics,
        NULL AS memory_type,
        NULL AS memory_key,
        NULL AS memory_value,
        1 AS sort_order
    FROM recent_messages
    
    UNION ALL
    
    SELECT 
        'summary' AS type,
        NULL AS role,
        summary AS content,
        NOW() AS created_at,
        summary,
        key_topics,
        NULL AS memory_type,
        NULL AS memory_key,
        NULL AS memory_value,
        2 AS sort_order
    FROM session_summary
    
    UNION ALL
    
    SELECT 
        'memory' AS type,
        NULL AS role,
        value AS content,
        NOW() AS created_at,
        NULL AS summary,
        NULL AS key_topics,
        memory_type,
        key AS memory_key,
        value AS memory_value,
        3 AS sort_order
    FROM user_memories
) combined_results
ORDER BY sort_order, created_at ASC;

-- name: DeleteOldSessions :exec
DELETE FROM chat_sessions 
WHERE user_id = $1 
AND last_activity < NOW() - INTERVAL '30 days';

-- name: GetSessionStats :one
SELECT 
    COUNT(*) AS total_messages,
    COUNT(DISTINCT conversation_messages.session_id) AS total_sessions,
    AVG(tokens_used) AS avg_tokens_per_message,
    SUM(tokens_used) AS total_tokens_used
FROM conversation_messages 
WHERE user_id = $1;
