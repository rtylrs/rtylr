-- Adds indexed Finance list/search support for invoices and expenses.
-- Duplicate-index errors are safe under the migration runner's --force mode.

ALTER TABLE finance_invoices ADD INDEX idx_finance_invoices_business_created (business_id, created_at);
ALTER TABLE finance_invoices ADD INDEX idx_finance_invoices_business_status_created (business_id, status, created_at);
ALTER TABLE finance_invoices ADD FULLTEXT INDEX idx_finance_invoices_search (invoice_number, client_name, client_email);

ALTER TABLE finance_expenses ADD INDEX idx_finance_expenses_business_date (business_id, expense_date);
ALTER TABLE finance_expenses ADD INDEX idx_finance_expenses_business_category_date (business_id, category_id, expense_date);
ALTER TABLE finance_expenses ADD FULLTEXT INDEX idx_finance_expenses_search (description, vendor);
