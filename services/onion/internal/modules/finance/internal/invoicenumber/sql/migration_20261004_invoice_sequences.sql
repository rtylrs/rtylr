CREATE TABLE IF NOT EXISTS finance_invoice_sequences (
 business_id BIGINT NOT NULL,
 prefix VARCHAR(32) NOT NULL,
 last_sequence BIGINT NOT NULL DEFAULT 0,
 PRIMARY KEY (business_id, prefix)
) ENGINE=InnoDB;
