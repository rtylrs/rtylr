-- Read-only views that define exactly what the AI (NL-to-SQL) layer may see.
-- Every view exposes owner_id so onion can enforce tenant isolation by requiring
-- `WHERE owner_id = <caller>` on every generated query. The AI's read-only DB
-- user is granted SELECT on these views ONLY — never on base tables — so columns
-- like PIN hashes, payment-processor ids, and other sensitive fields are never
-- reachable. Keep these lean and curated; add columns deliberately.

CREATE OR REPLACE VIEW ai_products AS
SELECT id, owner_id, name, sku, category_id, brand_id, supplier_id,
       cost_price, sale_price, compare_at_price, currency,
       stock_quantity, low_stock_threshold, reorder_quantity,
       is_active, track_inventory, lead_time_days, created_at, updated_at
FROM erp_products;

CREATE OR REPLACE VIEW ai_customers AS
SELECT id, owner_id, first_name, last_name, email, phone, city, country,
       loyalty_points, total_spent, visit_count, last_visit_at, loyalty_tier,
       is_active, created_at, updated_at
FROM crm_customers;

CREATE OR REPLACE VIEW ai_invoices AS
SELECT id, owner_id, invoice_number, client_name, status,
       issue_date, due_date, paid_date, currency,
       subtotal, discount_amount, tax_rate, tax_amount, total,
       amount_paid, amount_due, created_at, updated_at
FROM finance_invoices;

CREATE OR REPLACE VIEW ai_expenses AS
SELECT id, owner_id, category_id, description, amount, currency, expense_date,
       vendor, payment_method, is_tax_deductible, tax_amount, status,
       created_at, updated_at
FROM finance_expenses;

CREATE OR REPLACE VIEW ai_pos_orders AS
SELECT id, owner_id, order_number, type, status, customer_id, customer_name,
       staff_id, location_id, subtotal, discount_amount, tax_amount, total,
       tip_amount, payment_method, payment_status, created_at, paid_at
FROM pos_orders;

CREATE OR REPLACE VIEW ai_employees AS
SELECT id, owner_id, first_name, last_name, email, department_id, location_id,
       employment_type, status, salary, hourly_rate, overtime_rate,
       start_date, end_date, created_at, updated_at
FROM hr_employees;

CREATE OR REPLACE VIEW ai_tasks AS
SELECT id, owner_id, project_id, workspace_id, title, status, priority,
       assignee_id, due_date, start_date, completed_at,
       estimated_hours, actual_hours, created_at, updated_at
FROM flow_tasks;
