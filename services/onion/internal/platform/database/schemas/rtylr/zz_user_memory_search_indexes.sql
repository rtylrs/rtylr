-- Adds indexed user memory search support.

SET @sql := IF(
  (SELECT COUNT(*)
   FROM INFORMATION_SCHEMA.STATISTICS
   WHERE TABLE_SCHEMA = DATABASE()
     AND TABLE_NAME = 'user_memory'
     AND INDEX_NAME = 'idx_user_memory_search'),
  'SELECT 1',
  'ALTER TABLE user_memory ADD FULLTEXT INDEX idx_user_memory_search (`key`, value, memory_type)'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
