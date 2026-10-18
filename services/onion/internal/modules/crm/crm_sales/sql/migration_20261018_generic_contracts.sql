-- Contracts were shaped like rtylr's own subscriptions (plan Free/Pro/Scale,
-- seats, monthly MRR), so a business could not record what it actually sold.
-- `plan` becomes free text for what was sold, and a contract gains its value,
-- billing period, term and the CRM customer it belongs to. `mrr` stays as the
-- monthly equivalent so pipeline totals and older readers keep working.
-- Every step checks information_schema first, so a re-run is a no-op.

SET @plan_too_narrow := (
    SELECT COUNT(*) > 0 FROM information_schema.columns
    WHERE table_schema = DATABASE() AND table_name = 'crm_contracts'
      AND column_name = 'plan' AND character_maximum_length < 255
);
SET @plan_sql := IF(@plan_too_narrow,
    'ALTER TABLE `crm_contracts` MODIFY `plan` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT ''pro''',
    'SELECT 1');
PREPARE plan_statement FROM @plan_sql;
EXECUTE plan_statement;
DEALLOCATE PREPARE plan_statement;

SET @value_missing := (
    SELECT COUNT(*) = 0 FROM information_schema.columns
    WHERE table_schema = DATABASE() AND table_name = 'crm_contracts' AND column_name = 'value'
);
SET @value_sql := IF(@value_missing,
    'ALTER TABLE `crm_contracts`
       ADD COLUMN `value` decimal(12,2) NOT NULL DEFAULT ''0.00'' AFTER `seats`,
       ADD COLUMN `billing` enum(''one_time'',''monthly'',''yearly'') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT ''monthly'' AFTER `value`,
       ADD COLUMN `start_date` date DEFAULT NULL AFTER `billing`,
       ADD COLUMN `end_date` date DEFAULT NULL AFTER `start_date`,
       ADD COLUMN `customer_id` bigint DEFAULT NULL AFTER `end_date`',
    'SELECT 1');
PREPARE value_statement FROM @value_sql;
EXECUTE value_statement;
DEALLOCATE PREPARE value_statement;

-- Existing contracts were monthly subscriptions: their value is their MRR.
UPDATE `crm_contracts` SET `value` = `mrr`, `billing` = 'monthly'
WHERE `value` = 0 AND `mrr` > 0;

-- Link each existing contract to the customer its lead's contact became.
UPDATE `crm_contracts` AS contract
JOIN (
    SELECT contact.business_id, contact.lead_id,
           SUBSTRING_INDEX(GROUP_CONCAT(contact.customer_id ORDER BY contact.is_primary DESC, contact.id), ',', 1) AS customer_id
    FROM `crm_contacts` AS contact
    WHERE contact.customer_id IS NOT NULL
    GROUP BY contact.business_id, contact.lead_id
) AS linked ON linked.business_id = contract.business_id AND linked.lead_id = contract.lead_id
SET contract.customer_id = linked.customer_id
WHERE contract.customer_id IS NULL;
