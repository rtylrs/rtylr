-- Scope the AI analytics views by business_id, not owner_id.
--
-- ai_views.sql built every ai_* view around owner_id, and aiquery bound
-- `WHERE owner_id = <caller>` on every generated query. owner_id is shared by
-- every business of one account (tenant.ResolveOwnerID deliberately collapses
-- them), so a member granted access to business B1 could read B2's invoices,
-- customers, POS orders, employees (salary/hourly_rate/overtime_rate) and tasks
-- through Caret analyst mode.
--
-- Every base table below already carries a NOT NULL business_id with the same
-- semantics the rest of the backend scopes these tables by, so this is a
-- projection change only: no join, no derived column, and no subquery that
-- would put owner scoping back.
--
-- Named as its own zz_ file rather than an edit to ai_views.sql because
-- migrate.sh records applied files by name: editing an already-applied file is
-- a silent no-op on a migrated database. CREATE OR REPLACE VIEW is idempotent,
-- so a re-run is safe.
--
-- A row whose business_id is still the legacy 0 default stays invisible to the
-- AI layer — the same rows the rest of the backend already cannot see, because
-- it scopes these tables by business_id.

CREATE OR REPLACE VIEW ai_products AS
SELECT id, owner_id, business_id, name, sku, category_id, brand_id, supplier_id,
       cost_price, sale_price, compare_at_price, currency,
       stock_quantity, low_stock_threshold, reorder_quantity,
       is_active, track_inventory, lead_time_days, created_at, updated_at
FROM erp_products;

CREATE OR REPLACE VIEW ai_customers AS
SELECT id, owner_id, business_id, first_name, last_name, email, phone, city, country,
       loyalty_points, total_spent, visit_count, last_visit_at, loyalty_tier,
       is_active, created_at, updated_at
FROM crm_customers;

CREATE OR REPLACE VIEW ai_invoices AS
SELECT id, owner_id, business_id, invoice_number, client_name, status,
       issue_date, due_date, paid_date, currency,
       subtotal, discount_amount, tax_rate, tax_amount, total,
       amount_paid, amount_due, created_at, updated_at
FROM finance_invoices;

CREATE OR REPLACE VIEW ai_expenses AS
SELECT id, owner_id, business_id, category_id, description, amount, currency, expense_date,
       vendor, payment_method, is_tax_deductible, tax_amount, status,
       created_at, updated_at
FROM finance_expenses;

CREATE OR REPLACE VIEW ai_pos_orders AS
SELECT id, owner_id, business_id, order_number, type, status, customer_id, customer_name,
       staff_id, location_id, subtotal, discount_amount, tax_amount, total,
       tip_amount, payment_method, payment_status, created_at, paid_at
FROM pos_orders;

CREATE OR REPLACE VIEW ai_employees AS
SELECT id, owner_id, business_id, first_name, last_name, email, department_id, location_id,
       employment_type, status, salary, hourly_rate, overtime_rate,
       start_date, end_date, created_at, updated_at
FROM hr_employees;

CREATE OR REPLACE VIEW ai_tasks AS
SELECT id, owner_id, business_id, project_id, workspace_id, title, status, priority,
       assignee_id, due_date, start_date, completed_at,
       estimated_hours, actual_hours, created_at, updated_at
FROM flow_tasks;
