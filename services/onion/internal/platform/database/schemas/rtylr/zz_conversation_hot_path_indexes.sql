-- Keep chat history reads ordered within a session without a filesort.
SET @sql := IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'conversation_messages' AND INDEX_NAME = 'idx_conversation_messages_session_created') = 0,
  'ALTER TABLE conversation_messages ADD INDEX idx_conversation_messages_session_created (session_id, created_at)',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
