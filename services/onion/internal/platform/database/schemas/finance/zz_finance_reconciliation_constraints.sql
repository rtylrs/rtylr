-- Duplicate-index errors are safe under the migration runner's --force mode.
-- The unique key is the database-level guard against concurrent attempts to
-- match the same bank transaction within one business.
ALTER TABLE finance_reconciliation_matches
    ADD UNIQUE KEY uq_finance_reconciliation_business_bank (business_id, bank_feed_id);
