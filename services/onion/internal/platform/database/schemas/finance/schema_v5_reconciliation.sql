-- Phase 5 — reconciliation matches.
--
-- Records the outcome of matching a bank feed transaction against a
-- finance_invoices row or a finance_expenses row. The match is written
-- by POST /finance/reconciliation/match; GET /finance/reconciliation
-- reads the unmatched bank txns + unmatched invoices and suggests
-- pairings by amount + date proximity.
--
-- target_type discriminates between invoice and expense so a single
-- table captures both directions. owner_id is duplicated on every row
-- so every query can filter by it without a join.
--
-- migrate.sh runs with `mysql --force`, so a duplicate-table error on
-- re-runs is logged as a warning and migration continues. First run
-- applies cleanly.

CREATE TABLE IF NOT EXISTS finance_reconciliation_matches (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    bank_feed_id BIGINT NOT NULL,
    target_type VARCHAR(16) NOT NULL,
    target_id BIGINT NOT NULL,
    matched_by BIGINT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    KEY idx_finance_reconciliation_owner (owner_id),
    KEY idx_finance_reconciliation_bank (bank_feed_id),
    KEY idx_finance_reconciliation_target (target_type, target_id),
    KEY idx_finance_reconciliation_business_bank (business_id, bank_feed_id),
    KEY idx_finance_reconciliation_business_target (business_id, target_type, target_id)
);
