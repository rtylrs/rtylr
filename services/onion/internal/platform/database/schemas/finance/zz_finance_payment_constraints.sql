-- Duplicate-index errors are safe under the migration runner's --force mode.
-- NULL references remain exempt so legacy payments without an idempotency key
-- continue to be accepted; non-empty references are unique per invoice.
ALTER TABLE finance_payments
    ADD UNIQUE KEY uq_finance_payments_business_invoice_reference (business_id, invoice_id, reference);
