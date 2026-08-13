-- Business-leading indexes for HR tenant queries.
--
-- This file is intentionally placed in hr_time: the schema migration order
-- creates every HR table before hr_time runs. Each change is idempotent so it
-- is safe for both tracked migrations and the legacy --force migration path.
-- Existing owner-leading indexes remain available for owner-scoped callers.
-- Roll back an individual key only after checking live plans with:
--   DROP INDEX <index_name> ON <table_name>;

-- Legacy HR tables.
SET @sql := IF((SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hr_staff' AND INDEX_NAME = 'idx_hr_staff_business_status_created') = 0, 'ALTER TABLE hr_staff ADD INDEX idx_hr_staff_business_status_created (business_id, status, created_at, id)', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql := IF((SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hr_staff' AND INDEX_NAME = 'idx_hr_staff_business_user') = 0, 'ALTER TABLE hr_staff ADD INDEX idx_hr_staff_business_user (business_id, user_id)', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql := IF((SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hr_staff' AND INDEX_NAME = 'idx_hr_staff_business_department') = 0, 'ALTER TABLE hr_staff ADD INDEX idx_hr_staff_business_department (business_id, department_id)', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql := IF((SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hr_staff' AND INDEX_NAME = 'idx_hr_staff_business_email') = 0, 'ALTER TABLE hr_staff ADD INDEX idx_hr_staff_business_email (business_id, email)', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql := IF((SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hr_time_off_requests' AND INDEX_NAME = 'idx_hr_time_off_business_staff_dates') = 0, 'ALTER TABLE hr_time_off_requests ADD INDEX idx_hr_time_off_business_staff_dates (business_id, staff_id, status, start_date, end_date)', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql := IF((SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hr_payroll_records' AND INDEX_NAME = 'idx_hr_payroll_records_business_staff_period') = 0, 'ALTER TABLE hr_payroll_records ADD INDEX idx_hr_payroll_records_business_staff_period (business_id, staff_id, period_start, period_end)', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql := IF((SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hr_reviews' AND INDEX_NAME = 'idx_hr_reviews_business_staff_status') = 0, 'ALTER TABLE hr_reviews ADD INDEX idx_hr_reviews_business_staff_status (business_id, staff_id, status, created_at)', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql := IF((SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hr_one_on_ones' AND INDEX_NAME = 'idx_hr_one_on_ones_business_staff_scheduled') = 0, 'ALTER TABLE hr_one_on_ones ADD INDEX idx_hr_one_on_ones_business_staff_scheduled (business_id, staff_id, scheduled_at)', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql := IF((SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hr_documents' AND INDEX_NAME = 'idx_hr_documents_business_staff_expiry') = 0, 'ALTER TABLE hr_documents ADD INDEX idx_hr_documents_business_staff_expiry (business_id, staff_id, expiry_date)', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql := IF((SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hr_dashboard_widgets' AND INDEX_NAME = 'idx_hr_widgets_business_role') = 0, 'ALTER TABLE hr_dashboard_widgets ADD INDEX idx_hr_widgets_business_role (business_id, role_id)', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- HR core and transactional tables.
SET @sql := IF((SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hr_company_config' AND INDEX_NAME = 'idx_hr_company_config_business') = 0, 'ALTER TABLE hr_company_config ADD INDEX idx_hr_company_config_business (business_id)', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql := IF((SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hr_setup_steps' AND INDEX_NAME = 'idx_hr_setup_steps_business_key') = 0, 'ALTER TABLE hr_setup_steps ADD INDEX idx_hr_setup_steps_business_key (business_id, step_key)', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql := IF((SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hr_locations' AND INDEX_NAME = 'idx_hr_locations_business_active_name') = 0, 'ALTER TABLE hr_locations ADD INDEX idx_hr_locations_business_active_name (business_id, is_active, name)', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql := IF((SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hr_work_schedules' AND INDEX_NAME = 'idx_hr_work_schedules_business_default_name') = 0, 'ALTER TABLE hr_work_schedules ADD INDEX idx_hr_work_schedules_business_default_name (business_id, is_default, name)', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql := IF((SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hr_departments' AND INDEX_NAME = 'idx_hr_departments_business_location_name') = 0, 'ALTER TABLE hr_departments ADD INDEX idx_hr_departments_business_location_name (business_id, location_id, name)', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql := IF((SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hr_employees' AND INDEX_NAME = 'idx_hr_employees_business_status_created') = 0, 'ALTER TABLE hr_employees ADD INDEX idx_hr_employees_business_status_created (business_id, status, created_at, id)', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql := IF((SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hr_employees' AND INDEX_NAME = 'idx_hr_employees_business_user') = 0, 'ALTER TABLE hr_employees ADD INDEX idx_hr_employees_business_user (business_id, user_id)', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql := IF((SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hr_employees' AND INDEX_NAME = 'idx_hr_employees_business_department') = 0, 'ALTER TABLE hr_employees ADD INDEX idx_hr_employees_business_department (business_id, department_id)', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql := IF((SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hr_employees' AND INDEX_NAME = 'idx_hr_employees_business_location') = 0, 'ALTER TABLE hr_employees ADD INDEX idx_hr_employees_business_location (business_id, location_id)', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql := IF((SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hr_roles' AND INDEX_NAME = 'idx_hr_roles_business_system_name') = 0, 'ALTER TABLE hr_roles ADD INDEX idx_hr_roles_business_system_name (business_id, is_system, name)', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql := IF((SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hr_employee_roles' AND INDEX_NAME = 'idx_hr_employee_roles_business_employee_role') = 0, 'ALTER TABLE hr_employee_roles ADD INDEX idx_hr_employee_roles_business_employee_role (business_id, employee_id, role_id)', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql := IF((SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hr_approval_chains' AND INDEX_NAME = 'idx_hr_approval_chains_business_type') = 0, 'ALTER TABLE hr_approval_chains ADD INDEX idx_hr_approval_chains_business_type (business_id, request_type)', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql := IF((SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hr_approval_chain_levels' AND INDEX_NAME = 'idx_hr_approval_levels_business_chain_order') = 0, 'ALTER TABLE hr_approval_chain_levels ADD INDEX idx_hr_approval_levels_business_chain_order (business_id, approval_chain_id, level_order)', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql := IF((SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hr_employee_invites' AND INDEX_NAME = 'idx_hr_employee_invites_business_employee_expiry') = 0, 'ALTER TABLE hr_employee_invites ADD INDEX idx_hr_employee_invites_business_employee_expiry (business_id, employee_id, expires_at)', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql := IF((SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hr_leave_balances' AND INDEX_NAME = 'idx_hr_leave_balances_business_employee_type_year') = 0, 'ALTER TABLE hr_leave_balances ADD INDEX idx_hr_leave_balances_business_employee_type_year (business_id, employee_id, leave_type_id, year)', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql := IF((SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hr_attendance_records' AND INDEX_NAME = 'idx_hr_attendance_business_employee_date') = 0, 'ALTER TABLE hr_attendance_records ADD INDEX idx_hr_attendance_business_employee_date (business_id, employee_id, date)', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql := IF((SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hr_attendance_records' AND INDEX_NAME = 'idx_hr_attendance_business_location_date') = 0, 'ALTER TABLE hr_attendance_records ADD INDEX idx_hr_attendance_business_location_date (business_id, location_id, date)', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql := IF((SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hr_attendance_config' AND INDEX_NAME = 'idx_hr_attendance_config_business') = 0, 'ALTER TABLE hr_attendance_config ADD INDEX idx_hr_attendance_config_business (business_id)', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- Leave, onboarding, overtime, payroll, performance, shift, and time paths.
SET @sql := IF((SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hr_leave_types' AND INDEX_NAME = 'idx_hr_leave_types_business_active_name') = 0, 'ALTER TABLE hr_leave_types ADD INDEX idx_hr_leave_types_business_active_name (business_id, is_active, name)', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql := IF((SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hr_leave_requests' AND INDEX_NAME = 'idx_hr_leave_requests_business_employee_status_dates') = 0, 'ALTER TABLE hr_leave_requests ADD INDEX idx_hr_leave_requests_business_employee_status_dates (business_id, employee_id, status, start_date, end_date)', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql := IF((SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hr_leave_requests' AND INDEX_NAME = 'idx_hr_leave_requests_business_status_dates') = 0, 'ALTER TABLE hr_leave_requests ADD INDEX idx_hr_leave_requests_business_status_dates (business_id, status, start_date, end_date)', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql := IF((SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hr_onboarding_checklist_templates' AND INDEX_NAME = 'idx_hr_onboarding_templates_business_name') = 0, 'ALTER TABLE hr_onboarding_checklist_templates ADD INDEX idx_hr_onboarding_templates_business_name (business_id, name)', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql := IF((SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hr_onboarding_checklist_items' AND INDEX_NAME = 'idx_hr_onboarding_items_business_template') = 0, 'ALTER TABLE hr_onboarding_checklist_items ADD INDEX idx_hr_onboarding_items_business_template (business_id, template_id)', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql := IF((SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hr_onboarding_employee_checklists' AND INDEX_NAME = 'idx_hr_onboarding_employee_checklists_business_employee_status') = 0, 'ALTER TABLE hr_onboarding_employee_checklists ADD INDEX idx_hr_onboarding_employee_checklists_business_employee_status (business_id, employee_id, status)', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql := IF((SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hr_onboarding_employee_checklist_items' AND INDEX_NAME = 'idx_hr_onboarding_employee_items_business_checklist_status_due') = 0, 'ALTER TABLE hr_onboarding_employee_checklist_items ADD INDEX idx_hr_onboarding_employee_items_business_checklist_status_due (business_id, employee_checklist_id, status, due_date)', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql := IF((SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hr_overtime_requests' AND INDEX_NAME = 'idx_hr_overtime_business_employee_date_status') = 0, 'ALTER TABLE hr_overtime_requests ADD INDEX idx_hr_overtime_business_employee_date_status (business_id, employee_id, date, status)', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql := IF((SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hr_overtime_requests' AND INDEX_NAME = 'idx_hr_overtime_business_status_date') = 0, 'ALTER TABLE hr_overtime_requests ADD INDEX idx_hr_overtime_business_status_date (business_id, status, date)', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql := IF((SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hr_overtime_config' AND INDEX_NAME = 'idx_hr_overtime_config_business') = 0, 'ALTER TABLE hr_overtime_config ADD INDEX idx_hr_overtime_config_business (business_id)', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql := IF((SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hr_payroll_config' AND INDEX_NAME = 'idx_hr_payroll_config_business') = 0, 'ALTER TABLE hr_payroll_config ADD INDEX idx_hr_payroll_config_business (business_id)', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql := IF((SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hr_payroll_runs' AND INDEX_NAME = 'idx_hr_payroll_runs_business_status_created') = 0, 'ALTER TABLE hr_payroll_runs ADD INDEX idx_hr_payroll_runs_business_status_created (business_id, status, created_at)', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql := IF((SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hr_payroll_entries' AND INDEX_NAME = 'idx_hr_payroll_entries_business_run_employee') = 0, 'ALTER TABLE hr_payroll_entries ADD INDEX idx_hr_payroll_entries_business_run_employee (business_id, payroll_run_id, employee_id)', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql := IF((SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hr_payroll_entries' AND INDEX_NAME = 'idx_hr_payroll_entries_business_employee') = 0, 'ALTER TABLE hr_payroll_entries ADD INDEX idx_hr_payroll_entries_business_employee (business_id, employee_id)', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql := IF((SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hr_performance_review_cycles' AND INDEX_NAME = 'idx_hr_performance_cycles_business_status_period') = 0, 'ALTER TABLE hr_performance_review_cycles ADD INDEX idx_hr_performance_cycles_business_status_period (business_id, status, period_start, period_end)', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql := IF((SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hr_performance_reviews' AND INDEX_NAME = 'idx_hr_performance_reviews_business_employee_cycle_status') = 0, 'ALTER TABLE hr_performance_reviews ADD INDEX idx_hr_performance_reviews_business_employee_cycle_status (business_id, employee_id, review_cycle_id, status)', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql := IF((SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hr_performance_review_goals' AND INDEX_NAME = 'idx_hr_performance_goals_business_employee_cycle_status') = 0, 'ALTER TABLE hr_performance_review_goals ADD INDEX idx_hr_performance_goals_business_employee_cycle_status (business_id, employee_id, review_cycle_id, status)', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql := IF((SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hr_shift_assignments' AND INDEX_NAME = 'idx_hr_shift_assignments_business_employee_date') = 0, 'ALTER TABLE hr_shift_assignments ADD INDEX idx_hr_shift_assignments_business_employee_date (business_id, employee_id, date)', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql := IF((SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hr_shift_assignments' AND INDEX_NAME = 'idx_hr_shift_assignments_business_shift_date') = 0, 'ALTER TABLE hr_shift_assignments ADD INDEX idx_hr_shift_assignments_business_shift_date (business_id, shift_id, date)', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql := IF((SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hr_time_entries' AND INDEX_NAME = 'idx_hr_time_entries_business_employee_date') = 0, 'ALTER TABLE hr_time_entries ADD INDEX idx_hr_time_entries_business_employee_date (business_id, employee_id, date)', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql := IF((SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hr_time_entries' AND INDEX_NAME = 'idx_hr_time_entries_business_status_billable_date') = 0, 'ALTER TABLE hr_time_entries ADD INDEX idx_hr_time_entries_business_status_billable_date (business_id, status, is_billable, date)', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql := IF((SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hr_timesheets' AND INDEX_NAME = 'idx_hr_timesheets_business_employee_period') = 0, 'ALTER TABLE hr_timesheets ADD INDEX idx_hr_timesheets_business_employee_period (business_id, employee_id, period_start)', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql := IF((SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hr_timesheets' AND INDEX_NAME = 'idx_hr_timesheets_business_status_period') = 0, 'ALTER TABLE hr_timesheets ADD INDEX idx_hr_timesheets_business_status_period (business_id, status, period_start, period_end)', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @sql := IF((SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'hr_timesheet_approval_actions' AND INDEX_NAME = 'idx_hr_approval_actions_business_timesheet_created') = 0, 'ALTER TABLE hr_timesheet_approval_actions ADD INDEX idx_hr_approval_actions_business_timesheet_created (business_id, timesheet_id, created_at)', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
