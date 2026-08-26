-- Duplicate-index errors are safe under the migration runner's --force mode.
-- The unique key is the database-level guard against concurrent attempts to
-- match the same bank transaction within one business.
ALTER TABLE finance_reconciliation_matches
    ADD UNIQUE KEY uq_finance_reconciliation_business_bank (business_id, bank_feed_id);

-- Keep target reservations separate from the polymorphic match rows so this
-- migration can be applied without deleting or rewriting historical duplicate
-- matches. INSERT IGNORE preserves one reservation for each target and makes
-- all future writes race-safe at the database boundary.
CREATE TABLE IF NOT EXISTS finance_reconciliation_targets (
    business_id BIGINT NOT NULL,
    target_type VARCHAR(16) NOT NULL,
    target_id BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (business_id, target_type, target_id)
);

INSERT IGNORE INTO finance_reconciliation_targets (business_id, target_type, target_id)
SELECT business_id, target_type, target_id
FROM finance_reconciliation_matches;
