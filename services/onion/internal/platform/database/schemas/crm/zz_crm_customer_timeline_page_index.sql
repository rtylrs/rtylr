ALTER TABLE crm_customer_timeline
    ADD INDEX idx_crm_customer_timeline_page (business_id, customer_id, created_at, id);
