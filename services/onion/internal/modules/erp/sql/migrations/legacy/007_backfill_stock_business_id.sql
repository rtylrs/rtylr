-- Keep the denormalized tenant key populated on all unified stock rows.
-- Legacy migrations used business_id = 0; stock consumers scope by business_id,
-- so those rows were invisible to tenant-scoped reads and idempotency checks.
-- The canonical tenant key for these tables is owner_id.
UPDATE stock_locations
SET business_id = owner_id
WHERE business_id = 0;

UPDATE stock_levels
SET business_id = owner_id
WHERE business_id = 0;

UPDATE stock_ledger
SET business_id = owner_id
WHERE business_id = 0;

-- A payment retry must not be able to append the same stock movement twice.
-- MySQL permits multiple NULLs in a unique key, so unreferenced/manual rows
-- remain valid while referenced movements become idempotent per tenant/type.
ALTER TABLE stock_ledger
ADD UNIQUE KEY uq_ledger_reference (owner_id, reference_type, reference_id, entry_type);
