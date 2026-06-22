-- Adds indexed user memory search support.
-- Duplicate-index errors are safe under the migration runner's --force mode.

ALTER TABLE user_memory ADD FULLTEXT INDEX idx_user_memory_search (`key`, value, memory_type);
