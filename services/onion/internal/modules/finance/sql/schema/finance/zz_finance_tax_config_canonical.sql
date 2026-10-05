-- Canonical tax-config model shared by the tax and invoices modules.

-- Normalize the legacy tax_name/tax_rate columns first. Several rows can now
-- share (business_id, name, type) — the old table only guaranteed one row per
-- owner_id — so the default flag is handled separately below.
UPDATE finance_tax_config
SET name = CASE WHEN name = '' THEN COALESCE(NULLIF(tax_name, ''), 'Default tax') ELSE name END,
    rate = CASE WHEN rate = 0 AND tax_rate IS NOT NULL THEN tax_rate ELSE rate END
WHERE is_default = FALSE OR name = '';

-- Enforce the one-default-per-business invariant the tax module relies on.
-- A business that already has several defaults keeps its lowest id; one that
-- has configs but no default (invoices can create non-default rows) gets its
-- lowest id promoted.
UPDATE finance_tax_config
SET is_default = FALSE
WHERE is_default = TRUE
  AND id NOT IN (
      SELECT id FROM (
          SELECT MIN(id) AS id
          FROM finance_tax_config first_defaults
          WHERE first_defaults.is_default = TRUE
          GROUP BY first_defaults.business_id
      ) AS canonical_defaults
  );

UPDATE finance_tax_config
SET is_default = TRUE
WHERE COALESCE(is_default, FALSE) = FALSE
  AND id IN (
      SELECT id FROM (
          SELECT MIN(eligible.id) AS id
          FROM finance_tax_config eligible
          WHERE COALESCE(eligible.is_default, FALSE) = FALSE
            AND NOT EXISTS (
                SELECT 1 FROM (
                    SELECT DISTINCT business_id FROM finance_tax_config WHERE is_default = TRUE
                ) existing_defaults
                WHERE existing_defaults.business_id = eligible.business_id
            )
          GROUP BY eligible.business_id
      ) AS first_rows
  );

-- The unique key below cannot be added over the current duplicates. Following
-- zz_finance_invoice_number_unique.sql, every non-canonical duplicate is
-- suffixed with its own id (keeping the rename collision-free) instead of
-- dropping rows: the default row is canonical, otherwise the lowest id is.
UPDATE finance_tax_config
SET name = CONCAT(name, '-', id)
WHERE COALESCE(is_default, FALSE) = FALSE
  AND id IN (
      SELECT id FROM (
          SELECT dup.id
          FROM finance_tax_config dup
          WHERE EXISTS (
              SELECT 1 FROM finance_tax_config canonical
              WHERE canonical.business_id = dup.business_id
                AND canonical.name = dup.name
                AND canonical.type = dup.type
                AND canonical.id <> dup.id
                AND (canonical.is_default = TRUE OR canonical.id < dup.id)
          )
      ) AS duplicates
  )
  AND NOT EXISTS (
      SELECT 1 FROM INFORMATION_SCHEMA.STATISTICS
      WHERE TABLE_SCHEMA = DATABASE()
        AND TABLE_NAME = 'finance_tax_config'
        AND INDEX_NAME = 'uq_finance_tax_config_business_name_type'
  );


  


  
