-- Cross-domain performance indexes (2026-06). Every tenant query in these
-- domains was full-scanning — no index on owner_id/business_id or FK columns.
-- Pure ADD INDEX: non-destructive, applied to the live DB. ERP is already
-- indexed (migrations/001) and HR is intentionally EXCLUDED (owned by another
-- dev — its index recommendations are flagged for them, not applied here).
-- Kept under schemas/_perf/ (outside the sqlc-scanned domain dirs) as the record.

-- ── finance ──────────────────────────────────────────────────────────────
ALTER TABLE finance_invoices
  ADD INDEX idx_finance_invoices_owner (owner_id),
  ADD INDEX idx_finance_invoices_owner_status (owner_id, status),
  ADD INDEX idx_finance_invoices_owner_created (owner_id, created_at);
ALTER TABLE finance_expenses
  ADD INDEX idx_finance_expenses_owner (owner_id),
  ADD INDEX idx_finance_expenses_owner_category (owner_id, category_id);
ALTER TABLE finance_expense_categories ADD INDEX idx_finance_expense_categories_owner (owner_id);
ALTER TABLE finance_payments
  ADD INDEX idx_finance_payments_owner (owner_id),
  ADD INDEX idx_finance_payments_invoice (invoice_id);
ALTER TABLE finance_invoice_items ADD INDEX idx_finance_invoice_items_invoice (invoice_id);
ALTER TABLE finance_bills
  ADD INDEX idx_finance_bills_owner (owner_id),
  ADD INDEX idx_finance_bills_owner_status (owner_id, status);
ALTER TABLE finance_quotes
  ADD INDEX idx_finance_quotes_owner (owner_id),
  ADD INDEX idx_finance_quotes_owner_status (owner_id, status);
ALTER TABLE finance_recurring_invoices ADD INDEX idx_finance_recurring_invoices_owner (owner_id);
ALTER TABLE finance_bank_accounts ADD INDEX idx_finance_bank_accounts_owner (owner_id);
ALTER TABLE finance_journal_entries
  ADD INDEX idx_finance_journal_entries_owner (owner_id),
  ADD INDEX idx_finance_journal_entries_owner_status (owner_id, status);
ALTER TABLE finance_journal_lines
  ADD INDEX idx_finance_journal_lines_entry (entry_id),
  ADD INDEX idx_finance_journal_lines_account (account_id);

-- ── support ──────────────────────────────────────────────────────────────
ALTER TABLE support_tickets
  ADD INDEX idx_support_tickets_business (business_id),
  ADD INDEX idx_support_tickets_business_status (business_id, status),
  ADD INDEX idx_support_tickets_business_updated (business_id, updated_at);
ALTER TABLE support_messages
  ADD INDEX idx_support_messages_ticket (ticket_id),
  ADD INDEX idx_support_messages_ticket_created (ticket_id, created_at);
ALTER TABLE support_customers ADD INDEX idx_support_customers_business (business_id);
ALTER TABLE support_articles ADD INDEX idx_support_articles_business (business_id);
ALTER TABLE support_canned_responses ADD INDEX idx_support_canned_business (business_id);

-- ── crm ──────────────────────────────────────────────────────────────────
ALTER TABLE crm_customers
  ADD INDEX idx_crm_customers_owner (owner_id),
  ADD INDEX idx_crm_customers_owner_active (owner_id, is_active),
  ADD INDEX idx_crm_customers_owner_created (owner_id, created_at);
ALTER TABLE crm_loyalty_transactions
  ADD INDEX idx_crm_loyalty_tx_owner (owner_id),
  ADD INDEX idx_crm_loyalty_tx_customer (owner_id, customer_id);
ALTER TABLE crm_segments ADD INDEX idx_crm_segments_owner (owner_id);
ALTER TABLE crm_campaigns
  ADD INDEX idx_crm_campaigns_owner (owner_id),
  ADD INDEX idx_crm_campaigns_owner_status (owner_id, status);
ALTER TABLE crm_campaign_sends
  ADD INDEX idx_crm_campaign_sends_campaign (campaign_id),
  ADD INDEX idx_crm_campaign_sends_customer (customer_id);

-- ── store ────────────────────────────────────────────────────────────────
ALTER TABLE store_sites ADD INDEX idx_store_sites_owner (owner_id);
ALTER TABLE store_products
  ADD INDEX idx_store_products_owner (owner_id),
  ADD INDEX idx_store_products_owner_published (owner_id, is_published);
ALTER TABLE store_categories ADD INDEX idx_store_categories_owner (owner_id);
ALTER TABLE store_orders
  ADD INDEX idx_store_orders_owner (owner_id),
  ADD INDEX idx_store_orders_owner_status (owner_id, status),
  ADD INDEX idx_store_orders_owner_created (owner_id, created_at);
ALTER TABLE store_order_items ADD INDEX idx_store_order_items_order (order_id);
ALTER TABLE store_checkouts ADD INDEX idx_store_checkouts_owner (owner_id);
ALTER TABLE store_shipping_methods ADD INDEX idx_store_shipping_owner (owner_id);
ALTER TABLE store_payment_methods ADD INDEX idx_store_payment_methods_owner (owner_id);
ALTER TABLE store_themes ADD INDEX idx_store_themes_owner (owner_id);

-- ── booking ──────────────────────────────────────────────────────────────
ALTER TABLE booking_appointments
  ADD INDEX idx_booking_appts_owner (owner_id),
  ADD INDEX idx_booking_appts_owner_status (owner_id, status),
  ADD INDEX idx_booking_appts_owner_start (owner_id, start_time);
ALTER TABLE booking_services ADD INDEX idx_booking_services_owner (owner_id);
ALTER TABLE booking_resources ADD INDEX idx_booking_resources_owner (owner_id);
ALTER TABLE booking_waitlist ADD INDEX idx_booking_waitlist_owner (owner_id);
ALTER TABLE booking_reminders
  ADD INDEX idx_booking_reminders_owner (owner_id),
  ADD INDEX idx_booking_reminders_appt (appointment_id);

-- ── flow ─────────────────────────────────────────────────────────────────
ALTER TABLE flow_workspaces ADD INDEX idx_flow_workspaces_owner (owner_id);
ALTER TABLE flow_projects
  ADD INDEX idx_flow_projects_workspace_owner (workspace_id, owner_id),
  ADD INDEX idx_flow_projects_owner (owner_id);
ALTER TABLE flow_tasks
  ADD INDEX idx_flow_tasks_project (project_id),
  ADD INDEX idx_flow_tasks_workspace_owner (workspace_id, owner_id),
  ADD INDEX idx_flow_tasks_owner (owner_id),
  ADD INDEX idx_flow_tasks_assignee (assignee_id),
  ADD INDEX idx_flow_tasks_parent (parent_task_id),
  ADD INDEX idx_flow_tasks_owner_status (owner_id, status);
ALTER TABLE flow_task_comments ADD INDEX idx_flow_comments_task (task_id);
ALTER TABLE flow_task_activity ADD INDEX idx_flow_activity_task (task_id);
ALTER TABLE flow_task_attachments ADD INDEX idx_flow_attachments_task (task_id);
ALTER TABLE flow_time_logs
  ADD INDEX idx_flow_time_logs_task (task_id),
  ADD INDEX idx_flow_time_logs_owner (owner_id);
ALTER TABLE flow_milestones
  ADD INDEX idx_flow_milestones_owner (owner_id),
  ADD INDEX idx_flow_milestones_project (project_id);
ALTER TABLE flow_cycles ADD INDEX idx_flow_cycles_owner (owner_id);
ALTER TABLE flow_views ADD INDEX idx_flow_views_owner_workspace (owner_id, workspace_id);

-- ── pos (supplemental — base v2 indexes already exist for some) ───────────
ALTER TABLE pos_loyalty_transactions ADD INDEX idx_pos_loyalty_tx_business_customer (business_id, customer_id);
ALTER TABLE pos_store_credit_transactions ADD INDEX idx_pos_credit_tx_business_customer (business_id, customer_id);
ALTER TABLE pos_cash_movements ADD INDEX idx_pos_cash_movements_business_session (business_id, session_id);
ALTER TABLE pos_floors ADD INDEX idx_pos_floors_business_branch (business_id, branch_id);
ALTER TABLE pos_terminals ADD INDEX idx_pos_terminals_business_branch (business_id, branch_id);
