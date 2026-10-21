-- Dashboard activity (crm_sales dashboard.sql): calls per day and per rep read
-- crm_touchpoints by type and time, and qualified leads per rep read crm_leads
-- by conversion time. Deals won read idx_crm_deals_biz_status. Each index is
-- added only when missing, so a re-run is a no-op.

SET @touchpoints_type_missing := NOT EXISTS (
    SELECT 1 FROM information_schema.statistics
    WHERE table_schema = DATABASE() AND table_name = 'crm_touchpoints'
      AND index_name = 'idx_crm_touchpoints_biz_type'
);
SET @touchpoints_type_sql := IF(@touchpoints_type_missing,
    'ALTER TABLE `crm_touchpoints` ADD KEY `idx_crm_touchpoints_biz_type` (`business_id`,`type`,`occurred_at`)',
    'SELECT 1');
PREPARE touchpoints_type_statement FROM @touchpoints_type_sql;
EXECUTE touchpoints_type_statement;
DEALLOCATE PREPARE touchpoints_type_statement;

SET @leads_converted_missing := NOT EXISTS (
    SELECT 1 FROM information_schema.statistics
    WHERE table_schema = DATABASE() AND table_name = 'crm_leads'
      AND index_name = 'idx_crm_leads_biz_converted'
);
SET @leads_converted_sql := IF(@leads_converted_missing,
    'ALTER TABLE `crm_leads` ADD KEY `idx_crm_leads_biz_converted` (`business_id`,`converted_at`)',
    'SELECT 1');
PREPARE leads_converted_statement FROM @leads_converted_sql;
EXECUTE leads_converted_statement;
DEALLOCATE PREPARE leads_converted_statement;
