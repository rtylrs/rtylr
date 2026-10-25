-- Flow work days, their actions and work types belong to a business, not to its
-- owner: keys move to business_id and owner_id goes nullable so this build can
-- stop writing it. A later release drops the column.

SET @apply := (SELECT COUNT(*) = 0 FROM information_schema.statistics WHERE table_schema = DATABASE() AND table_name = 'flow_work_days' AND index_name = 'uk_work_day_business');
SET @stmt := IF(@apply, 'ALTER TABLE flow_work_days ADD UNIQUE KEY uk_work_day_business (business_id, employee_id, date)', 'SELECT 1');
PREPARE work_time_stmt FROM @stmt;
EXECUTE work_time_stmt;
DEALLOCATE PREPARE work_time_stmt;

SET @apply := (SELECT COUNT(*) = 0 FROM information_schema.statistics WHERE table_schema = DATABASE() AND table_name = 'flow_work_days' AND index_name = 'idx_work_days_business_status');
SET @stmt := IF(@apply, 'ALTER TABLE flow_work_days ADD KEY idx_work_days_business_status (business_id, status, date)', 'SELECT 1');
PREPARE work_time_stmt FROM @stmt;
EXECUTE work_time_stmt;
DEALLOCATE PREPARE work_time_stmt;

SET @apply := (SELECT COUNT(*) > 0 FROM information_schema.statistics WHERE table_schema = DATABASE() AND table_name = 'flow_work_days' AND index_name = 'uk_work_day');
SET @stmt := IF(@apply, 'ALTER TABLE flow_work_days DROP INDEX uk_work_day', 'SELECT 1');
PREPARE work_time_stmt FROM @stmt;
EXECUTE work_time_stmt;
DEALLOCATE PREPARE work_time_stmt;

SET @apply := (SELECT COUNT(*) > 0 FROM information_schema.statistics WHERE table_schema = DATABASE() AND table_name = 'flow_work_days' AND index_name = 'idx_work_days_tenant_employee_date');
SET @stmt := IF(@apply, 'ALTER TABLE flow_work_days DROP INDEX idx_work_days_tenant_employee_date', 'SELECT 1');
PREPARE work_time_stmt FROM @stmt;
EXECUTE work_time_stmt;
DEALLOCATE PREPARE work_time_stmt;

SET @apply := (SELECT COUNT(*) > 0 FROM information_schema.statistics WHERE table_schema = DATABASE() AND table_name = 'flow_work_days' AND index_name = 'idx_work_days_tenant_status');
SET @stmt := IF(@apply, 'ALTER TABLE flow_work_days DROP INDEX idx_work_days_tenant_status', 'SELECT 1');
PREPARE work_time_stmt FROM @stmt;
EXECUTE work_time_stmt;
DEALLOCATE PREPARE work_time_stmt;

SET @apply := (SELECT COUNT(*) > 0 FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'flow_work_days' AND column_name = 'owner_id' AND is_nullable = 'NO');
SET @stmt := IF(@apply, 'ALTER TABLE flow_work_days MODIFY owner_id bigint DEFAULT NULL', 'SELECT 1');
PREPARE work_time_stmt FROM @stmt;
EXECUTE work_time_stmt;
DEALLOCATE PREPARE work_time_stmt;

SET @apply := (SELECT COUNT(*) = 0 FROM information_schema.statistics WHERE table_schema = DATABASE() AND table_name = 'flow_work_day_actions' AND index_name = 'idx_work_day_actions_business_day');
SET @stmt := IF(@apply, 'ALTER TABLE flow_work_day_actions ADD KEY idx_work_day_actions_business_day (business_id, work_day_id, created_at)', 'SELECT 1');
PREPARE work_time_stmt FROM @stmt;
EXECUTE work_time_stmt;
DEALLOCATE PREPARE work_time_stmt;

SET @apply := (SELECT COUNT(*) > 0 FROM information_schema.statistics WHERE table_schema = DATABASE() AND table_name = 'flow_work_day_actions' AND index_name = 'idx_work_day_actions_day');
SET @stmt := IF(@apply, 'ALTER TABLE flow_work_day_actions DROP INDEX idx_work_day_actions_day', 'SELECT 1');
PREPARE work_time_stmt FROM @stmt;
EXECUTE work_time_stmt;
DEALLOCATE PREPARE work_time_stmt;

SET @apply := (SELECT COUNT(*) > 0 FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'flow_work_day_actions' AND column_name = 'owner_id' AND is_nullable = 'NO');
SET @stmt := IF(@apply, 'ALTER TABLE flow_work_day_actions MODIFY owner_id bigint DEFAULT NULL', 'SELECT 1');
PREPARE work_time_stmt FROM @stmt;
EXECUTE work_time_stmt;
DEALLOCATE PREPARE work_time_stmt;

SET @apply := (SELECT COUNT(*) = 0 FROM information_schema.statistics WHERE table_schema = DATABASE() AND table_name = 'flow_work_types' AND index_name = 'idx_work_types_business_workspace');
SET @stmt := IF(@apply, 'ALTER TABLE flow_work_types ADD KEY idx_work_types_business_workspace (business_id, workspace_id, archived, sort_order)', 'SELECT 1');
PREPARE work_time_stmt FROM @stmt;
EXECUTE work_time_stmt;
DEALLOCATE PREPARE work_time_stmt;

SET @apply := (SELECT COUNT(*) = 0 FROM information_schema.statistics WHERE table_schema = DATABASE() AND table_name = 'flow_work_types' AND index_name = 'idx_work_types_business_parent');
SET @stmt := IF(@apply, 'ALTER TABLE flow_work_types ADD KEY idx_work_types_business_parent (business_id, parent_id)', 'SELECT 1');
PREPARE work_time_stmt FROM @stmt;
EXECUTE work_time_stmt;
DEALLOCATE PREPARE work_time_stmt;

SET @apply := (SELECT COUNT(*) > 0 FROM information_schema.statistics WHERE table_schema = DATABASE() AND table_name = 'flow_work_types' AND index_name = 'idx_work_types_tenant_workspace');
SET @stmt := IF(@apply, 'ALTER TABLE flow_work_types DROP INDEX idx_work_types_tenant_workspace', 'SELECT 1');
PREPARE work_time_stmt FROM @stmt;
EXECUTE work_time_stmt;
DEALLOCATE PREPARE work_time_stmt;

SET @apply := (SELECT COUNT(*) > 0 FROM information_schema.statistics WHERE table_schema = DATABASE() AND table_name = 'flow_work_types' AND index_name = 'idx_work_types_parent');
SET @stmt := IF(@apply, 'ALTER TABLE flow_work_types DROP INDEX idx_work_types_parent', 'SELECT 1');
PREPARE work_time_stmt FROM @stmt;
EXECUTE work_time_stmt;
DEALLOCATE PREPARE work_time_stmt;

SET @apply := (SELECT COUNT(*) > 0 FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'flow_work_types' AND column_name = 'owner_id' AND is_nullable = 'NO');
SET @stmt := IF(@apply, 'ALTER TABLE flow_work_types MODIFY owner_id bigint DEFAULT NULL', 'SELECT 1');
PREPARE work_time_stmt FROM @stmt;
EXECUTE work_time_stmt;
DEALLOCATE PREPARE work_time_stmt;
