CREATE TABLE IF NOT EXISTS `conversation_messages` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `session_id` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `user_id` int NOT NULL,
  `role` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `content` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `tokens_used` int DEFAULT NULL,
  `metadata` json DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_conversation_messages_session_id` (`session_id`),
  KEY `idx_conversation_messages_user_id` (`user_id`),
  KEY `idx_conversation_messages_created_at` (`created_at`),
  KEY `idx_conversation_messages_session_created` (`session_id`,`created_at`),
  FULLTEXT KEY `idx_conversation_messages_content_fts` (`content`),
  CONSTRAINT `conversation_messages_chk_1` CHECK ((`role` in (_utf8mb4'user',_utf8mb4'assistant',_utf8mb4'system')))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;





