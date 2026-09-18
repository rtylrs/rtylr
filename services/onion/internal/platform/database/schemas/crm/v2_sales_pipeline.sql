-- Sales Pipeline add-on tables (CRM module extension)
-- All tables are tenant-scoped via owner_id.

CREATE TABLE IF NOT EXISTS crm_leads (
  id            BIGINT AUTO_INCREMENT PRIMARY KEY,
  owner_id      BIGINT NOT NULL,
  business_id   BIGINT NOT NULL DEFAULT 0,
  company_name  VARCHAR(255) NOT NULL,
  industry      VARCHAR(100) NULL,
  size          VARCHAR(50) NULL,
  source        VARCHAR(50) NOT NULL DEFAULT 'cold_email',
  stage         VARCHAR(50) NOT NULL DEFAULT 'cold',
  assigned_to   BIGINT NULL,
  mrr_potential DECIMAL(10,2) NULL,
  notes         TEXT NULL,
  created_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_crm_leads_owner (owner_id),
  INDEX idx_crm_leads_stage (owner_id, stage),
  INDEX idx_crm_leads_assigned (owner_id, assigned_to),
  INDEX idx_crm_leads_business (business_id)
);

CREATE TABLE IF NOT EXISTS crm_contacts (
  id         BIGINT AUTO_INCREMENT PRIMARY KEY,
  owner_id   BIGINT NOT NULL,
  business_id BIGINT NOT NULL DEFAULT 0,
  lead_id    BIGINT NOT NULL,
  customer_id BIGINT NULL,
  name       VARCHAR(255) NOT NULL,
  role       VARCHAR(100) NULL,
  email      VARCHAR(255) NULL,
  phone      VARCHAR(50) NULL,
  linkedin   VARCHAR(255) NULL,
  is_primary BOOLEAN NOT NULL DEFAULT FALSE,
  notes      TEXT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_crm_contacts_owner (owner_id),
  INDEX idx_crm_contacts_lead (lead_id),
  INDEX idx_crm_contacts_customer (owner_id, customer_id),
  INDEX idx_crm_contacts_business (business_id)
);

CREATE TABLE IF NOT EXISTS crm_touchpoints (
  id             BIGINT AUTO_INCREMENT PRIMARY KEY,
  owner_id       BIGINT NOT NULL,
  business_id    BIGINT NOT NULL DEFAULT 0,
  lead_id        BIGINT NOT NULL,
  contact_id     BIGINT NULL,
  type           VARCHAR(50) NOT NULL DEFAULT 'call',
  outcome        VARCHAR(50) NOT NULL DEFAULT 'other',
  notes          TEXT NULL,
  next_follow_up DATE NULL,
  created_by     BIGINT NOT NULL,
  created_at     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_crm_touchpoints_owner (owner_id),
  INDEX idx_crm_touchpoints_lead (lead_id),
  INDEX idx_crm_touchpoints_business (business_id)
);

CREATE TABLE IF NOT EXISTS crm_contracts (
  id             BIGINT AUTO_INCREMENT PRIMARY KEY,
  owner_id       BIGINT NOT NULL,
  business_id    BIGINT NOT NULL DEFAULT 0,
  lead_id        BIGINT NOT NULL,
  plan           VARCHAR(50) NOT NULL DEFAULT 'pro',
  mrr            DECIMAL(10,2) NOT NULL DEFAULT 0,
  seats          INT NOT NULL DEFAULT 1,
  signed_at      DATE NULL,
  signed_by_name VARCHAR(255) NULL,
  closed_by      BIGINT NULL,
  notes          TEXT NULL,
  created_at     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_crm_contracts_owner (owner_id),
  INDEX idx_crm_contracts_lead (lead_id),
  INDEX idx_crm_contracts_business (business_id)
);
