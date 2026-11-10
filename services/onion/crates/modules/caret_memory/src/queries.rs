//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CreateChatSession` (Go `CreateChatSessionParams`).
pub struct CreateChatSessionParams {
    pub user_id: i32,
    pub session_id: String,
    pub title: Option<String>,
    pub metadata: String,
}

/// `CreateChatSession`: `execresult`.
pub async fn create_chat_session<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateChatSessionParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query(
        "INSERT INTO chat_sessions (user_id, session_id, title, metadata)\nVALUES (?, ?, ?, ?)",
    )
    .bind(params.user_id)
    .bind(&params.session_id)
    .bind(&params.title)
    .bind(&params.metadata)
    .execute(exec)
    .await
}

/// Params for `CreateConversationSummary` (Go `CreateConversationSummaryParams`).
pub struct CreateConversationSummaryParams {
    pub user_id: i32,
    pub session_id: String,
    pub summary: String,
    pub key_topics: String,
    pub sentiment: Option<String>,
    pub importance_score: Option<f64>,
}

/// `CreateConversationSummary`: `execresult`.
pub async fn create_conversation_summary<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateConversationSummaryParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO conversation_summaries (user_id, session_id, summary, key_topics, sentiment, importance_score)\nVALUES (?, ?, ?, ?, ?, ?)")
        .bind(params.user_id)
        .bind(&params.session_id)
        .bind(&params.summary)
        .bind(&params.key_topics)
        .bind(&params.sentiment)
        .bind(params.importance_score)
        .execute(exec)
        .await
}

/// Params for `CreateMessage` (Go `CreateMessageParams`).
pub struct CreateMessageParams {
    pub session_id: String,
    pub user_id: i32,
    pub role: String,
    pub content: String,
    pub tokens_used: Option<i32>,
    pub metadata: String,
}

/// `CreateMessage`: `execresult`.
pub async fn create_message<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateMessageParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO conversation_messages (session_id, user_id, role, content, tokens_used, metadata)\nVALUES (?, ?, ?, ?, ?, ?)")
        .bind(&params.session_id)
        .bind(params.user_id)
        .bind(&params.role)
        .bind(&params.content)
        .bind(params.tokens_used)
        .bind(&params.metadata)
        .execute(exec)
        .await
}

/// Params for `DeleteChatSessionByID` (Go `DeleteChatSessionByIDParams`).
pub struct DeleteChatSessionByIDParams {
    pub session_id: String,
    pub user_id: i32,
}

/// `DeleteChatSessionByID`: `execresult`.
pub async fn delete_chat_session_by_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteChatSessionByIDParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM chat_sessions WHERE session_id = ? AND user_id = ?")
        .bind(&params.session_id)
        .bind(params.user_id)
        .execute(exec)
        .await
}

/// Params for `DeleteSessionMessages` (Go `DeleteSessionMessagesParams`).
pub struct DeleteSessionMessagesParams {
    pub session_id: String,
    pub user_id: i32,
}

/// `DeleteSessionMessages`: `exec`.
pub async fn delete_session_messages<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteSessionMessagesParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM conversation_messages WHERE session_id = ? AND user_id = ?")
        .bind(&params.session_id)
        .bind(params.user_id)
        .execute(exec)
        .await
}

/// Params for `DeleteSessionSummaries` (Go `DeleteSessionSummariesParams`).
pub struct DeleteSessionSummariesParams {
    pub session_id: String,
    pub user_id: i32,
}

/// `DeleteSessionSummaries`: `exec`.
pub async fn delete_session_summaries<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteSessionSummariesParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM conversation_summaries WHERE session_id = ? AND user_id = ?")
        .bind(&params.session_id)
        .bind(params.user_id)
        .execute(exec)
        .await
}

/// Params for `GetChatSession` (Go `GetChatSessionParams`).
pub struct GetChatSessionParams {
    pub session_id: String,
    pub user_id: i32,
}

/// `GetChatSession`: `one`.
pub async fn get_chat_session<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetChatSessionParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT\n    id, user_id, session_id, title, summary, metadata,\n    created_at, updated_at, last_activity\nFROM chat_sessions\nWHERE session_id = ? AND user_id = ?")
        .bind(&params.session_id)
        .bind(params.user_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetContextForSession` (Go `GetContextForSessionParams`).
pub struct GetContextForSessionParams {
    pub session_id: String,
    pub session_id_2: String,
    pub user_id: i32,
}

/// `GetContextForSession`: `many`.
pub async fn get_context_for_session<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetContextForSessionParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("WITH recent_messages AS (\n    SELECT id, session_id, user_id, role, content, tokens_used, metadata, created_at\n    FROM conversation_messages\n    WHERE conversation_messages.session_id = ?\n    ORDER BY created_at DESC\n    LIMIT 20\n),\nsession_summary AS (\n    SELECT summary, key_topics FROM conversation_summaries \n    WHERE conversation_summaries.session_id = ? \n    ORDER BY created_at DESC \n    LIMIT 1\n),\nuser_memories AS (\n    SELECT memory_type, `key`, value FROM user_memory\n    WHERE user_memory.user_id = ?\n    AND confidence > 0.7\n    ORDER BY confidence DESC \n    LIMIT 10\n)\nSELECT\n    type,\n    COALESCE(role, '') AS role,\n    COALESCE(content, '') AS content,\n    created_at,\n    COALESCE(summary, '') AS summary,\n    COALESCE(key_topics, '') AS key_topics,\n    COALESCE(memory_type, '') AS memory_type,\n    COALESCE(memory_key, '') AS memory_key,\n    COALESCE(memory_value, '') AS memory_value\nFROM (\n    SELECT \n        'message' AS type,\n        recent_messages.role,\n        recent_messages.content,\n        recent_messages.created_at,\n        NULL AS summary,\n        NULL AS key_topics,\n        NULL AS memory_type,\n        NULL AS memory_key,\n        NULL AS memory_value,\n        1 AS sort_order\n    FROM recent_messages\n    \n    UNION ALL\n    \n    SELECT \n        'summary' AS type,\n        NULL AS role,\n        summary AS content,\n        CURRENT_TIMESTAMP AS created_at,\n        summary,\n        key_topics,\n        NULL AS memory_type,\n        NULL AS memory_key,\n        NULL AS memory_value,\n        2 AS sort_order\n    FROM session_summary\n    \n    UNION ALL\n    \n    SELECT \n        'memory' AS type,\n        NULL AS role,\n        value AS content,\n        CURRENT_TIMESTAMP AS created_at,\n        NULL AS summary,\n        NULL AS key_topics,\n        memory_type,\n        `key` AS memory_key,\n        value AS memory_value,\n        3 AS sort_order\n    FROM user_memories\n) combined_results\nORDER BY sort_order, created_at ASC")
        .bind(&params.session_id)
        .bind(&params.session_id_2)
        .bind(params.user_id)
        .execute(exec)
        .await
}

/// Params for `GetConversationSummaryByID` (Go `GetConversationSummaryByIDParams`).
pub struct GetConversationSummaryByIDParams {
    pub id: i64,
}

/// `GetConversationSummaryByID`: `one`.
pub async fn get_conversation_summary_by_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetConversationSummaryByIDParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT\n    id, user_id, session_id, summary, key_topics,\n    sentiment, importance_score, created_at\nFROM conversation_summaries\nWHERE id = ?")
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetFirstUserMessageContent` (Go `GetFirstUserMessageContentParams`).
pub struct GetFirstUserMessageContentParams {
    pub session_id: String,
    pub user_id: i32,
}

/// `GetFirstUserMessageContent`: `one`.
pub async fn get_first_user_message_content<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetFirstUserMessageContentParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT content\nFROM conversation_messages\nWHERE session_id = ? AND user_id = ? AND role = 'user'\nORDER BY created_at ASC, id ASC\nLIMIT 1")
        .bind(&params.session_id)
        .bind(params.user_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetLatestAssistantMessageTime` (Go `GetLatestAssistantMessageTimeParams`).
pub struct GetLatestAssistantMessageTimeParams {
    pub session_id: String,
}

/// `GetLatestAssistantMessageTime`: `one`.
pub async fn get_latest_assistant_message_time<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetLatestAssistantMessageTimeParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT created_at FROM conversation_messages\nWHERE session_id = ? AND role = 'assistant'\nORDER BY id DESC\nLIMIT 1")
        .bind(&params.session_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetMessageByID` (Go `GetMessageByIDParams`).
pub struct GetMessageByIDParams {
    pub id: i64,
}

/// `GetMessageByID`: `one`.
pub async fn get_message_by_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetMessageByIDParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT\n    id, session_id, user_id, role, content,\n    tokens_used, metadata, created_at\nFROM conversation_messages\nWHERE id = ?")
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetSessionMessages` (Go `GetSessionMessagesParams`).
pub struct GetSessionMessagesParams {
    pub session_id: String,
}

/// `GetSessionMessages`: `many`.
pub async fn get_session_messages<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetSessionMessagesParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT\n    id, session_id, user_id, role, content,\n    tokens_used, metadata, created_at\nFROM conversation_messages\nWHERE session_id = ?\nORDER BY created_at ASC")
        .bind(&params.session_id)
        .execute(exec)
        .await
}

/// Params for `GetSessionStats` (Go `GetSessionStatsParams`).
pub struct GetSessionStatsParams {
    pub user_id: i32,
}

/// `GetSessionStats`: `one`.
pub async fn get_session_stats<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetSessionStatsParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT \n    COUNT(*) AS total_messages,\n    COUNT(DISTINCT conversation_messages.session_id) AS total_sessions,\n    AVG(tokens_used) AS avg_tokens_per_message,\n    SUM(tokens_used) AS total_tokens_used\nFROM conversation_messages \nWHERE user_id = ?")
        .bind(params.user_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetUserMemoryByKey` (Go `GetUserMemoryByKeyParams`).
pub struct GetUserMemoryByKeyParams {
    pub user_id: i32,
    pub memory_type: String,
    pub key: String,
}

/// `GetUserMemoryByKey`: `one`.
pub async fn get_user_memory_by_key<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetUserMemoryByKeyParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT\n    id, user_id, memory_type, `key`, value, confidence,\n    source_session, created_at, updated_at\nFROM user_memory\nWHERE user_id = ? AND memory_type = ? AND `key` = ?")
        .bind(params.user_id)
        .bind(&params.memory_type)
        .bind(&params.key)
        .fetch_optional(exec)
        .await
}

/// Params for `ListUserSessions` (Go `ListUserSessionsParams`).
pub struct ListUserSessionsParams {
    pub user_id: i32,
    pub limit: i32,
    pub offset: i32,
}

/// `ListUserSessions`: `many`.
pub async fn list_user_sessions<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListUserSessionsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT\n    id, user_id, session_id, title, summary, metadata,\n    created_at, updated_at, last_activity\nFROM chat_sessions\nWHERE user_id = ? AND session_id NOT LIKE 'buddy-%'\nORDER BY last_activity DESC\nLIMIT ? OFFSET ?")
        .bind(params.user_id)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `SearchUserMemory` (Go `SearchUserMemoryParams`).
pub struct SearchUserMemoryParams {
    pub user_id: i32,
    pub memory_type: String,
    pub key: String,
}

/// `SearchUserMemory`: `many`.
pub async fn search_user_memory<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &SearchUserMemoryParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT\n    id, user_id, memory_type, `key`, value, confidence,\n    source_session, created_at, updated_at\nFROM user_memory\nWHERE user_id = ?\nAND (\n    memory_type = ? OR\n    `key` = ?\n)\nORDER BY confidence DESC")
        .bind(params.user_id)
        .bind(&params.memory_type)
        .bind(&params.key)
        .execute(exec)
        .await
}

/// Params for `StoreUserMemory` (Go `StoreUserMemoryParams`).
pub struct StoreUserMemoryParams {
    pub user_id: i32,
    pub memory_type: String,
    pub key: String,
    pub value: String,
    pub confidence: Option<f64>,
    pub source_session: Option<String>,
}

/// `StoreUserMemory`: `exec`.
pub async fn store_user_memory<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &StoreUserMemoryParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO user_memory (user_id, memory_type, `key`, value, confidence, source_session)\nVALUES (?, ?, ?, ?, ?, ?)\nON DUPLICATE KEY UPDATE\n    value = VALUES(value),\n    confidence = VALUES(confidence),\n    source_session = VALUES(source_session),\n    updated_at = CURRENT_TIMESTAMP")
        .bind(params.user_id)
        .bind(&params.memory_type)
        .bind(&params.key)
        .bind(&params.value)
        .bind(params.confidence)
        .bind(&params.source_session)
        .execute(exec)
        .await
}

/// Params for `UpdateChatSession` (Go `UpdateChatSessionParams`).
pub struct UpdateChatSessionParams {
    pub title: Option<String>,
    pub summary: Option<String>,
    pub metadata: String,
    pub session_id: String,
    pub user_id: i32,
}

/// `UpdateChatSession`: `exec`.
pub async fn update_chat_session<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateChatSessionParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE chat_sessions \nSET title = ?, summary = ?, metadata = ?, last_activity = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP\nWHERE session_id = ? AND user_id = ?")
        .bind(&params.title)
        .bind(&params.summary)
        .bind(&params.metadata)
        .bind(&params.session_id)
        .bind(params.user_id)
        .execute(exec)
        .await
}
