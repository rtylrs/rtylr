-- name: CreateChatSession :execresult
INSERT INTO chat_sessions (user_id, session_id, title, metadata)
VALUES (?, ?, ?, ?);

-- name: GetChatSession :one
SELECT
    id, user_id, session_id, title, summary, metadata,
    created_at, updated_at, last_activity
FROM chat_sessions
WHERE session_id = ? AND user_id = ?;

-- name: UpdateChatSession :exec
UPDATE chat_sessions 
SET title = ?, summary = ?, metadata = ?, last_activity = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
WHERE session_id = ? AND user_id = ?;

-- name: ListUserSessions :many
SELECT
    id, user_id, session_id, title, summary, metadata,
    created_at, updated_at, last_activity
FROM chat_sessions
WHERE user_id = ? AND session_id NOT LIKE 'buddy-%'
ORDER BY last_activity DESC
LIMIT ? OFFSET ?;

-- name: CreateMessage :execresult
INSERT INTO conversation_messages (session_id, user_id, role, content, tokens_used, metadata)
VALUES (?, ?, ?, ?, ?, ?);

-- name: GetMessageByID :one
SELECT
    id, session_id, user_id, role, content,
    tokens_used, metadata, created_at
FROM conversation_messages
WHERE id = ?;

-- name: GetSessionMessages :many
SELECT
    id, session_id, user_id, role, content,
    tokens_used, metadata, created_at
FROM conversation_messages
WHERE session_id = ?
ORDER BY created_at ASC;

-- name: CreateConversationSummary :execresult
INSERT INTO conversation_summaries (user_id, session_id, summary, key_topics, sentiment, importance_score)
VALUES (?, ?, ?, ?, ?, ?);

-- name: GetConversationSummaryByID :one
SELECT
    id, user_id, session_id, summary, key_topics,
    sentiment, importance_score, created_at
FROM conversation_summaries
WHERE id = ?;

-- name: StoreUserMemory :exec
INSERT INTO user_memory (user_id, memory_type, `key`, value, confidence, source_session)
VALUES (?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
    value = VALUES(value),
    confidence = VALUES(confidence),
    source_session = VALUES(source_session),
    updated_at = CURRENT_TIMESTAMP;

-- name: GetUserMemoryByKey :one
SELECT
    id, user_id, memory_type, `key`, value, confidence,
    source_session, created_at, updated_at
FROM user_memory
WHERE user_id = ? AND memory_type = ? AND `key` = ?;

-- name: SearchUserMemory :many
SELECT
    id, user_id, memory_type, `key`, value, confidence,
    source_session, created_at, updated_at
FROM user_memory
WHERE user_id = ?
AND (
    memory_type = ? OR
    `key` = ?
)
ORDER BY confidence DESC;

-- name: GetContextForSession :many
WITH recent_messages AS (
    SELECT id, session_id, user_id, role, content, tokens_used, metadata, created_at
    FROM conversation_messages
    WHERE conversation_messages.session_id = ?
    ORDER BY created_at DESC
    LIMIT 20
),
session_summary AS (
    SELECT summary, key_topics FROM conversation_summaries 
    WHERE conversation_summaries.session_id = ? 
    ORDER BY created_at DESC 
    LIMIT 1
),
user_memories AS (
    SELECT memory_type, `key`, value FROM user_memory
    WHERE user_memory.user_id = ?
    AND confidence > 0.7
    ORDER BY confidence DESC 
    LIMIT 10
)
SELECT
    type,
    COALESCE(role, '') AS role,
    COALESCE(content, '') AS content,
    created_at,
    COALESCE(summary, '') AS summary,
    COALESCE(key_topics, '') AS key_topics,
    COALESCE(memory_type, '') AS memory_type,
    COALESCE(memory_key, '') AS memory_key,
    COALESCE(memory_value, '') AS memory_value
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
        CURRENT_TIMESTAMP AS created_at,
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
        CURRENT_TIMESTAMP AS created_at,
        NULL AS summary,
        NULL AS key_topics,
        memory_type,
        `key` AS memory_key,
        value AS memory_value,
        3 AS sort_order
    FROM user_memories
) combined_results
ORDER BY sort_order, created_at ASC;

-- name: GetSessionStats :one
SELECT 
    COUNT(*) AS total_messages,
    COUNT(DISTINCT conversation_messages.session_id) AS total_sessions,
    AVG(tokens_used) AS avg_tokens_per_message,
    SUM(tokens_used) AS total_tokens_used
FROM conversation_messages 
WHERE user_id = ?;

-- name: DeleteSessionMessages :exec
DELETE FROM conversation_messages WHERE session_id = ? AND user_id = ?;

-- name: DeleteSessionSummaries :exec
DELETE FROM conversation_summaries WHERE session_id = ? AND user_id = ?;

-- name: DeleteChatSessionByID :execresult
DELETE FROM chat_sessions WHERE session_id = ? AND user_id = ?;

-- name: GetFirstUserMessageContent :one
SELECT content
FROM conversation_messages
WHERE session_id = ? AND user_id = ? AND role = 'user'
ORDER BY created_at ASC, id ASC
LIMIT 1;
