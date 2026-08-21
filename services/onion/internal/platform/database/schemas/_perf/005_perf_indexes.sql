-- Cross-domain performance indexes (2026-06). Idempotent supplemental indexes.
-- Each index is added only when absent so the schema migrator can be rerun
-- safely on existing environments.

DROP PROCEDURE IF EXISTS add_perf_index;
DELIMITER //
CREATE PROCEDURE add_perf_index(IN p_table VARCHAR(64), IN p_index VARCHAR(64), IN p_columns VARCHAR(255))
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM INFORMATION_SCHEMA.STATISTICS
        WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = p_table AND INDEX_NAME = p_index
    ) THEN
        SET @stmt := CONCAT('ALTER TABLE `', p_table, '` ADD INDEX `', p_index, '` (', p_columns, ')');
        PREPARE add_index_stmt FROM @stmt;
        EXECUTE add_index_stmt;
        DEALLOCATE PREPARE add_index_stmt;
    END IF;
END//
DELIMITER ;

CALL add_perf_index('finance_invoices', 'idx_finance_invoices_owner', 'owner_id');
CALL add_perf_index('finance_invoices', 'idx_finance_invoices_owner_status', 'owner_id, status');
CALL add_perf_index('finance_invoices', 'idx_finance_invoices_owner_created', 'owner_id, created_at');
CALL add_perf_index('finance_expenses', 'idx_finance_expenses_owner', 'owner_id');
CALL add_perf_index('finance_expenses', 'idx_finance_expenses_owner_category', 'owner_id, category_id');
CALL add_perf_index('finance_expense_categories', 'idx_finance_expense_categories_owner', 'owner_id');
CALL add_perf_index('finance_payments', 'idx_finance_payments_owner', 'owner_id');
CALL add_perf_index('finance_payments', 'idx_finance_payments_invoice', 'invoice_id');
CALL add_perf_index('finance_invoice_items', 'idx_finance_invoice_items_invoice', 'invoice_id');
CALL add_perf_index('finance_bills', 'idx_finance_bills_owner', 'owner_id');
CALL add_perf_index('finance_bills', 'idx_finance_bills_owner_status', 'owner_id, status');
CALL add_perf_index('finance_quotes', 'idx_finance_quotes_owner', 'owner_id');
CALL add_perf_index('finance_quotes', 'idx_finance_quotes_owner_status', 'owner_id, status');
CALL add_perf_index('finance_recurring_invoices', 'idx_finance_recurring_invoices_owner', 'owner_id');
CALL add_perf_index('finance_bank_accounts', 'idx_finance_bank_accounts_owner', 'owner_id');
CALL add_perf_index('finance_journal_entries', 'idx_finance_journal_entries_owner', 'owner_id');
CALL add_perf_index('finance_journal_entries', 'idx_finance_journal_entries_owner_status', 'owner_id, status');
CALL add_perf_index('finance_journal_lines', 'idx_finance_journal_lines_entry', 'entry_id');
CALL add_perf_index('finance_journal_lines', 'idx_finance_journal_lines_account', 'account_id');

CALL add_perf_index('support_tickets', 'idx_support_tickets_business', 'business_id');
CALL add_perf_index('support_tickets', 'idx_support_tickets_business_status', 'business_id, status');
CALL add_perf_index('support_tickets', 'idx_support_tickets_business_updated', 'business_id, updated_at');
CALL add_perf_index('support_messages', 'idx_support_messages_ticket', 'ticket_id');
CALL add_perf_index('support_messages', 'idx_support_messages_ticket_created', 'ticket_id, created_at');
CALL add_perf_index('support_customers', 'idx_support_customers_business', 'business_id');
CALL add_perf_index('support_articles', 'idx_support_articles_business', 'business_id');
CALL add_perf_index('support_canned_responses', 'idx_support_canned_business', 'business_id');

CALL add_perf_index('crm_customers', 'idx_crm_customers_owner', 'owner_id');
CALL add_perf_index('crm_customers', 'idx_crm_customers_owner_active', 'owner_id, is_active');
CALL add_perf_index('crm_customers', 'idx_crm_customers_owner_created', 'owner_id, created_at');
CALL add_perf_index('crm_loyalty_transactions', 'idx_crm_loyalty_tx_owner', 'owner_id');
CALL add_perf_index('crm_loyalty_transactions', 'idx_crm_loyalty_tx_customer', 'owner_id, customer_id');
CALL add_perf_index('crm_segments', 'idx_crm_segments_owner', 'owner_id');
CALL add_perf_index('crm_campaigns', 'idx_crm_campaigns_owner', 'owner_id');
CALL add_perf_index('crm_campaigns', 'idx_crm_campaigns_owner_status', 'owner_id, status');
CALL add_perf_index('crm_campaign_sends', 'idx_crm_campaign_sends_campaign', 'campaign_id');
CALL add_perf_index('crm_campaign_sends', 'idx_crm_campaign_sends_customer', 'customer_id');

CALL add_perf_index('store_sites', 'idx_store_sites_owner', 'owner_id');
CALL add_perf_index('store_products', 'idx_store_products_owner', 'owner_id');
CALL add_perf_index('store_products', 'idx_store_products_owner_published', 'owner_id, is_published');
CALL add_perf_index('store_categories', 'idx_store_categories_owner', 'owner_id');
CALL add_perf_index('store_orders', 'idx_store_orders_owner', 'owner_id');
CALL add_perf_index('store_orders', 'idx_store_orders_owner_status', 'owner_id, status');
CALL add_perf_index('store_orders', 'idx_store_orders_owner_created', 'owner_id, created_at');
CALL add_perf_index('store_order_items', 'idx_store_order_items_order', 'order_id');
CALL add_perf_index('store_checkouts', 'idx_store_checkouts_owner', 'owner_id');
CALL add_perf_index('store_shipping_methods', 'idx_store_shipping_owner', 'owner_id');
CALL add_perf_index('store_payment_methods', 'idx_store_payment_methods_owner', 'owner_id');
CALL add_perf_index('store_themes', 'idx_store_themes_owner', 'owner_id');

CALL add_perf_index('booking_appointments', 'idx_booking_appts_owner', 'owner_id');
CALL add_perf_index('booking_appointments', 'idx_booking_appts_owner_status', 'owner_id, status');
CALL add_perf_index('booking_appointments', 'idx_booking_appts_owner_start', 'owner_id, start_time');
CALL add_perf_index('booking_services', 'idx_booking_services_owner', 'owner_id');
CALL add_perf_index('booking_resources', 'idx_booking_resources_owner', 'owner_id');
CALL add_perf_index('booking_waitlist', 'idx_booking_waitlist_owner', 'owner_id');
CALL add_perf_index('booking_reminders', 'idx_booking_reminders_owner', 'owner_id');
CALL add_perf_index('booking_reminders', 'idx_booking_reminders_appt', 'appointment_id');

CALL add_perf_index('flow_workspaces', 'idx_flow_workspaces_owner', 'owner_id');
CALL add_perf_index('flow_projects', 'idx_flow_projects_workspace_owner', 'workspace_id, owner_id');
CALL add_perf_index('flow_projects', 'idx_flow_projects_owner', 'owner_id');
CALL add_perf_index('flow_tasks', 'idx_flow_tasks_project', 'project_id');
CALL add_perf_index('flow_tasks', 'idx_flow_tasks_workspace_owner', 'workspace_id, owner_id');
CALL add_perf_index('flow_tasks', 'idx_flow_tasks_owner', 'owner_id');
CALL add_perf_index('flow_tasks', 'idx_flow_tasks_assignee', 'assignee_id');
CALL add_perf_index('flow_tasks', 'idx_flow_tasks_parent', 'parent_task_id');
CALL add_perf_index('flow_tasks', 'idx_flow_tasks_owner_status', 'owner_id, status');
CALL add_perf_index('flow_task_comments', 'idx_flow_comments_task', 'task_id');
CALL add_perf_index('flow_task_activity', 'idx_flow_activity_task', 'task_id');
CALL add_perf_index('flow_task_attachments', 'idx_flow_attachments_task', 'task_id');
CALL add_perf_index('flow_time_logs', 'idx_flow_time_logs_task', 'task_id');
CALL add_perf_index('flow_time_logs', 'idx_flow_time_logs_owner', 'owner_id');
CALL add_perf_index('flow_milestones', 'idx_flow_milestones_owner', 'owner_id');
CALL add_perf_index('flow_milestones', 'idx_flow_milestones_project', 'project_id');
CALL add_perf_index('flow_cycles', 'idx_flow_cycles_owner', 'owner_id');
CALL add_perf_index('flow_views', 'idx_flow_views_owner_workspace', 'owner_id, workspace_id');

CALL add_perf_index('pos_loyalty_transactions', 'idx_pos_loyalty_tx_business_customer', 'business_id, customer_id');
CALL add_perf_index('pos_store_credit_transactions', 'idx_pos_credit_tx_business_customer', 'business_id, customer_id');
CALL add_perf_index('pos_cash_movements', 'idx_pos_cash_movements_business_session', 'business_id, session_id');
CALL add_perf_index('pos_floors', 'idx_pos_floors_business_branch', 'business_id, branch_id');
CALL add_perf_index('pos_terminals', 'idx_pos_terminals_business_branch', 'business_id, branch_id');

DROP PROCEDURE add_perf_index;
