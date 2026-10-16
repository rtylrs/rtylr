-- Business-scoped projection of this module's tables for the AI analyst (aiquery).
-- Published here so the owner controls the columns aiquery may read.
-- owner_id is shared by every business of an account, so scope by business_id.
-- Kept as its own zz_ file: migrate.sh records applied files by name.

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
