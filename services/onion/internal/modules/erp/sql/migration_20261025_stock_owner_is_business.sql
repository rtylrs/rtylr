-- ERP rows carry the business ID in owner_id. A few writers stored the owner's
-- user ID instead: the POS menu, POS stock and online orders. Rewrite those
-- rows so stock and products resolve by business. IGNORE skips a row whose
-- business already has the same key; that duplicate stays unreachable rather
-- than blocking the release. Production had none of these rows.

UPDATE IGNORE stock_locations SET owner_id = business_id WHERE business_id > 0 AND owner_id <> business_id;
UPDATE IGNORE stock_levels SET owner_id = business_id WHERE business_id > 0 AND owner_id <> business_id;
UPDATE IGNORE stock_ledger SET owner_id = business_id WHERE business_id > 0 AND owner_id <> business_id;
UPDATE inventory_cost_layers SET owner_id = business_id WHERE business_id > 0 AND owner_id <> business_id;
UPDATE erp_products SET owner_id = business_id WHERE business_id > 0 AND owner_id <> business_id;
