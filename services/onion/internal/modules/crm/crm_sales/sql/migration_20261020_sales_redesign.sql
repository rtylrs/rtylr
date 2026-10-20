-- Sales redesign (docs/crm-model.md): leads carry the prospect's contact
-- details and a working status, qualified work moves to deals in per-business
-- pipelines, follow-ups become tasks, and imports are batched.
--
-- Every step checks information_schema or the data it writes, so a re-run
-- after a stopped migration is a no-op. Legacy columns (`stage`,
-- `mrr_potential`) and `crm_contacts` stay for the rollout and are only read.

CREATE TABLE IF NOT EXISTS `crm_companies` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `business_id` bigint NOT NULL,
  `owner_id` bigint NOT NULL,
  `name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `website` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `industry` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `size` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `phone` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `city` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `country` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `assigned_to` bigint DEFAULT NULL,
  `notes` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_crm_companies_biz_name` (`business_id`,`name`,`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `crm_pipelines` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `business_id` bigint NOT NULL,
  `name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `position` int NOT NULL DEFAULT '0',
  `default_marker` tinyint DEFAULT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_crm_pipelines_default` (`business_id`,`default_marker`),
  KEY `idx_crm_pipelines_business` (`business_id`,`position`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `crm_pipeline_stages` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `business_id` bigint NOT NULL,
  `pipeline_id` bigint NOT NULL,
  `name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `position` int NOT NULL DEFAULT '0',
  `probability` tinyint unsigned NOT NULL DEFAULT '0',
  `kind` enum('open','won','lost') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'open',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_crm_pipeline_stages_pipeline` (`business_id`,`pipeline_id`,`position`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `crm_deals` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `business_id` bigint NOT NULL,
  `owner_id` bigint NOT NULL,
  `pipeline_id` bigint NOT NULL,
  `stage_id` bigint NOT NULL,
  `title` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `lead_id` bigint DEFAULT NULL,
  `customer_id` bigint DEFAULT NULL,
  `company_id` bigint DEFAULT NULL,
  `contact_name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `company_name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `value` decimal(12,2) NOT NULL DEFAULT '0.00',
  `billing` enum('one_time','monthly','yearly') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'one_time',
  `expected_close_date` date DEFAULT NULL,
  `assigned_to` bigint DEFAULT NULL,
  `source` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `status` enum('open','won','lost') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'open',
  `lost_reason` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `closed_at` timestamp NULL DEFAULT NULL,
  `notes` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
  `last_activity_at` timestamp NULL DEFAULT NULL,
  `next_follow_up` date DEFAULT NULL,
  `stage_changed_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_crm_deals_stage` (`business_id`,`pipeline_id`,`stage_id`,`created_at`,`id`),
  KEY `idx_crm_deals_biz_created` (`business_id`,`created_at`,`id`),
  KEY `idx_crm_deals_biz_status` (`business_id`,`status`,`closed_at`),
  KEY `idx_crm_deals_biz_rep` (`business_id`,`assigned_to`),
  KEY `idx_crm_deals_biz_close` (`business_id`,`expected_close_date`,`id`),
  KEY `idx_crm_deals_lead` (`business_id`,`lead_id`),
  KEY `idx_crm_deals_customer` (`business_id`,`customer_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `crm_tasks` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `business_id` bigint NOT NULL,
  `owner_id` bigint NOT NULL,
  `lead_id` bigint DEFAULT NULL,
  `deal_id` bigint DEFAULT NULL,
  `title` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `type` enum('call','email','meeting','todo') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'todo',
  `due_date` date NOT NULL,
  `assigned_to` bigint DEFAULT NULL,
  `completed_at` timestamp NULL DEFAULT NULL,
  `created_by` bigint NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_crm_tasks_open` (`business_id`,`completed_at`,`due_date`,`id`),
  KEY `idx_crm_tasks_rep` (`business_id`,`assigned_to`,`completed_at`,`due_date`),
  KEY `idx_crm_tasks_lead` (`business_id`,`lead_id`),
  KEY `idx_crm_tasks_deal` (`business_id`,`deal_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `crm_import_batches` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `business_id` bigint NOT NULL,
  `created_by` bigint NOT NULL,
  `file_name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `total_rows` int NOT NULL DEFAULT '0',
  `created_count` int NOT NULL DEFAULT '0',
  `updated_count` int NOT NULL DEFAULT '0',
  `skipped_count` int NOT NULL DEFAULT '0',
  `failed_count` int NOT NULL DEFAULT '0',
  `status` enum('running','completed','undone') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'running',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `completed_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_crm_import_batches_business` (`business_id`,`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Leads: contact details, status, follow-up rollups and list indexes.
SET @leads_missing := (
    SELECT COUNT(*) = 0 FROM information_schema.columns
    WHERE table_schema = DATABASE() AND table_name = 'crm_leads' AND column_name = 'status'
);
SET @leads_sql := IF(@leads_missing,
    'ALTER TABLE `crm_leads`
       ADD COLUMN `first_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL AFTER `business_id`,
       ADD COLUMN `last_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL AFTER `first_name`,
       ADD COLUMN `job_title` varchar(150) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL AFTER `last_name`,
       ADD COLUMN `email` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL AFTER `job_title`,
       ADD COLUMN `phone` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL AFTER `email`,
       ADD COLUMN `phone_digits` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL AFTER `phone`,
       MODIFY `company_name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '''',
       ADD COLUMN `website` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL AFTER `company_name`,
       ADD COLUMN `linkedin` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL AFTER `website`,
       ADD COLUMN `city` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL AFTER `size`,
       ADD COLUMN `country` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL AFTER `city`,
       MODIFY `source` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '''',
       ADD COLUMN `status` enum(''new'',''attempting'',''connected'',''nurture'',''qualified'',''disqualified'') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT ''new'' AFTER `source`,
       ADD COLUMN `disqualify_reason` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL AFTER `status`,
       ADD COLUMN `value` decimal(12,2) DEFAULT NULL AFTER `assigned_to`,
       ADD COLUMN `billing` enum(''one_time'',''monthly'',''yearly'') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT ''one_time'' AFTER `value`,
       ADD COLUMN `attempt_count` int NOT NULL DEFAULT ''0'' AFTER `notes`,
       ADD COLUMN `last_activity_at` timestamp NULL DEFAULT NULL AFTER `attempt_count`,
       ADD COLUMN `next_follow_up` date DEFAULT NULL AFTER `last_activity_at`,
       ADD COLUMN `customer_id` bigint DEFAULT NULL AFTER `next_follow_up`,
       ADD COLUMN `company_id` bigint DEFAULT NULL AFTER `customer_id`,
       ADD COLUMN `deal_id` bigint DEFAULT NULL AFTER `company_id`,
       ADD COLUMN `converted_at` timestamp NULL DEFAULT NULL AFTER `deal_id`,
       ADD COLUMN `import_batch_id` bigint DEFAULT NULL AFTER `converted_at`,
       ADD KEY `idx_crm_leads_biz_created` (`business_id`,`created_at`,`id`),
       ADD KEY `idx_crm_leads_biz_status` (`business_id`,`status`,`created_at`,`id`),
       ADD KEY `idx_crm_leads_biz_rep` (`business_id`,`assigned_to`,`created_at`,`id`),
       ADD KEY `idx_crm_leads_biz_follow_up` (`business_id`,`next_follow_up`,`id`),
       ADD KEY `idx_crm_leads_biz_activity` (`business_id`,`last_activity_at`,`id`),
       ADD KEY `idx_crm_leads_biz_company` (`business_id`,`company_name`,`id`),
       ADD KEY `idx_crm_leads_biz_email` (`business_id`,`email`),
       ADD KEY `idx_crm_leads_biz_phone` (`business_id`,`phone_digits`),
       ADD KEY `idx_crm_leads_biz_batch` (`business_id`,`import_batch_id`)',
    'SELECT 1');
PREPARE leads_statement FROM @leads_sql;
EXECUTE leads_statement;
DEALLOCATE PREPARE leads_statement;

-- The lead's primary (or first) contact becomes the lead's person.
UPDATE `crm_leads` l
JOIN (
    SELECT c.lead_id, c.business_id, TRIM(c.name) AS name, c.role, c.email, c.phone, c.linkedin, c.customer_id,
           ROW_NUMBER() OVER (PARTITION BY c.lead_id ORDER BY c.is_primary DESC, c.id) AS position
    FROM `crm_contacts` c
) c ON c.lead_id = l.id AND c.business_id = l.business_id AND c.position = 1
SET l.first_name = SUBSTRING_INDEX(c.name, ' ', 1),
    l.last_name = NULLIF(TRIM(SUBSTRING(c.name, CHAR_LENGTH(SUBSTRING_INDEX(c.name, ' ', 1)) + 1)), ''),
    l.job_title = NULLIF(c.role, ''),
    l.email = NULLIF(LOWER(TRIM(c.email)), ''),
    l.phone = NULLIF(TRIM(c.phone), ''),
    l.linkedin = NULLIF(c.linkedin, ''),
    l.customer_id = COALESCE(l.customer_id, c.customer_id),
    l.updated_at = l.updated_at
WHERE l.first_name IS NULL AND l.email IS NULL AND l.phone IS NULL;

UPDATE `crm_leads`
SET phone_digits = NULLIF(REGEXP_REPLACE(phone, '[^0-9]', ''), ''), updated_at = updated_at
WHERE phone IS NOT NULL AND phone_digits IS NULL;

UPDATE `crm_leads`
SET status = CASE stage
        WHEN 'contacted' THEN 'connected'
        WHEN 'lost' THEN 'disqualified'
        ELSE 'qualified'
    END,
    disqualify_reason = IF(stage = 'lost', 'Marked lost', disqualify_reason),
    updated_at = updated_at
WHERE status = 'new' AND stage IN ('contacted', 'demo_booked', 'demo_done', 'proposal_sent', 'contracted', 'lost');

UPDATE `crm_leads`
SET value = mrr_potential, billing = 'monthly', updated_at = updated_at
WHERE value IS NULL AND mrr_potential IS NOT NULL;

-- Leads that were past the demo stage are deals in the default pipeline. Stage
-- names and probabilities match application.DefaultStages.
INSERT IGNORE INTO `crm_pipelines` (business_id, name, position, default_marker)
SELECT DISTINCT business_id, 'Sales', 0, 1
FROM `crm_leads`
WHERE stage IN ('demo_booked', 'demo_done', 'proposal_sent', 'contracted') AND deal_id IS NULL;

INSERT INTO `crm_pipeline_stages` (business_id, pipeline_id, name, position, probability, kind)
SELECT p.business_id, p.id, s.name, s.position, s.probability, s.kind
FROM `crm_pipelines` p
JOIN (
    SELECT 'Qualified' AS name, 0 AS position, 10 AS probability, 'open' AS kind
    UNION ALL SELECT 'Meeting', 1, 25, 'open'
    UNION ALL SELECT 'Proposal', 2, 50, 'open'
    UNION ALL SELECT 'Negotiation', 3, 75, 'open'
    UNION ALL SELECT 'Won', 4, 100, 'won'
    UNION ALL SELECT 'Lost', 5, 0, 'lost'
) s
WHERE p.default_marker = 1
  AND NOT EXISTS (SELECT 1 FROM `crm_pipeline_stages` x WHERE x.pipeline_id = p.id AND x.business_id = p.business_id);

INSERT INTO `crm_deals` (business_id, owner_id, pipeline_id, stage_id, title, lead_id, customer_id,
                         contact_name, company_name, value, billing, assigned_to, source, status,
                         closed_at, notes, stage_changed_at, created_at)
SELECT l.business_id, l.owner_id, p.id, st.id,
       COALESCE(NULLIF(l.company_name, ''), NULLIF(TRIM(CONCAT_WS(' ', l.first_name, l.last_name)), ''), 'Deal'),
       l.id, l.customer_id, TRIM(CONCAT_WS(' ', l.first_name, l.last_name)), l.company_name,
       COALESCE(l.mrr_potential, 0), 'monthly', l.assigned_to, l.source,
       IF(l.stage = 'contracted', 'won', 'open'),
       IF(l.stage = 'contracted', COALESCE((SELECT MIN(c.created_at) FROM `crm_contracts` c
           WHERE c.lead_id = l.id AND c.business_id = l.business_id), l.updated_at), NULL),
       l.notes, l.updated_at, l.created_at
FROM `crm_leads` l
JOIN `crm_pipelines` p ON p.business_id = l.business_id AND p.default_marker = 1
JOIN `crm_pipeline_stages` st ON st.pipeline_id = p.id AND st.business_id = p.business_id
    AND st.name = CASE l.stage
        WHEN 'proposal_sent' THEN 'Proposal'
        WHEN 'contracted' THEN 'Won'
        ELSE 'Meeting'
    END
WHERE l.stage IN ('demo_booked', 'demo_done', 'proposal_sent', 'contracted')
  AND l.deal_id IS NULL
  AND NOT EXISTS (SELECT 1 FROM `crm_deals` d WHERE d.lead_id = l.id AND d.business_id = l.business_id);

UPDATE `crm_leads` l
JOIN `crm_deals` d ON d.lead_id = l.id AND d.business_id = l.business_id
SET l.deal_id = d.id, l.converted_at = COALESCE(l.converted_at, d.created_at), l.updated_at = l.updated_at
WHERE l.deal_id IS NULL;

-- Activities: any record can carry them, calls have a duration, and
-- `occurred_at` is when the touch happened rather than when it was logged.
SET @touchpoints_missing := (
    SELECT COUNT(*) = 0 FROM information_schema.columns
    WHERE table_schema = DATABASE() AND table_name = 'crm_touchpoints' AND column_name = 'deal_id'
);
SET @touchpoints_sql := IF(@touchpoints_missing,
    'ALTER TABLE `crm_touchpoints`
       MODIFY `lead_id` bigint DEFAULT NULL,
       ADD COLUMN `deal_id` bigint DEFAULT NULL AFTER `lead_id`,
       ADD COLUMN `customer_id` bigint DEFAULT NULL AFTER `deal_id`,
       MODIFY `outcome` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '''',
       ADD COLUMN `duration_seconds` int DEFAULT NULL AFTER `notes`,
       ADD COLUMN `occurred_at` timestamp NULL DEFAULT NULL AFTER `created_by`,
       ADD KEY `idx_crm_touchpoints_biz_lead` (`business_id`,`lead_id`,`occurred_at`,`id`),
       ADD KEY `idx_crm_touchpoints_biz_deal` (`business_id`,`deal_id`,`occurred_at`,`id`),
       ADD KEY `idx_crm_touchpoints_biz_rep` (`business_id`,`created_by`,`occurred_at`)',
    'SELECT 1');
PREPARE touchpoints_statement FROM @touchpoints_sql;
EXECUTE touchpoints_statement;
DEALLOCATE PREPARE touchpoints_statement;

UPDATE `crm_touchpoints` SET occurred_at = created_at WHERE occurred_at IS NULL;

SET @occurred_nullable := (
    SELECT COUNT(*) > 0 FROM information_schema.columns
    WHERE table_schema = DATABASE() AND table_name = 'crm_touchpoints'
      AND column_name = 'occurred_at' AND is_nullable = 'YES'
);
SET @occurred_sql := IF(@occurred_nullable,
    'ALTER TABLE `crm_touchpoints` MODIFY `occurred_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP',
    'SELECT 1');
PREPARE occurred_statement FROM @occurred_sql;
EXECUTE occurred_statement;
DEALLOCATE PREPARE occurred_statement;

UPDATE `crm_touchpoints` SET type = 'call' WHERE type = 'phone';
UPDATE `crm_touchpoints`
SET outcome = CASE outcome
        WHEN 'positive' THEN 'connected'
        WHEN 'neutral' THEN 'connected'
        WHEN 'negative' THEN 'not_interested'
        ELSE ''
    END
WHERE outcome IN ('positive', 'neutral', 'negative', 'other');

UPDATE `crm_leads` l
JOIN (
    SELECT lead_id, business_id, MAX(occurred_at) AS last_at, SUM(type = 'call') AS calls
    FROM `crm_touchpoints`
    WHERE lead_id IS NOT NULL
    GROUP BY lead_id, business_id
) t ON t.lead_id = l.id AND t.business_id = l.business_id
SET l.last_activity_at = t.last_at, l.attempt_count = t.calls, l.updated_at = l.updated_at
WHERE l.last_activity_at IS NULL;

-- A lead's latest follow-up date becomes an open call task.
INSERT INTO `crm_tasks` (business_id, owner_id, lead_id, deal_id, title, type, due_date, assigned_to, created_by, created_at)
SELECT l.business_id, l.owner_id, l.id, l.deal_id, 'Follow up', 'call', t.next_follow_up, l.assigned_to, t.created_by, t.created_at
FROM `crm_leads` l
JOIN (
    SELECT lead_id, business_id, next_follow_up, created_by, created_at,
           ROW_NUMBER() OVER (PARTITION BY lead_id ORDER BY created_at DESC, id DESC) AS position
    FROM `crm_touchpoints`
    WHERE lead_id IS NOT NULL
) t ON t.lead_id = l.id AND t.business_id = l.business_id AND t.position = 1
WHERE t.next_follow_up IS NOT NULL
  AND NOT EXISTS (SELECT 1 FROM `crm_tasks` x WHERE x.lead_id = l.id AND x.business_id = l.business_id);

UPDATE `crm_leads` l
JOIN (
    SELECT lead_id, business_id, MIN(due_date) AS due
    FROM `crm_tasks`
    WHERE completed_at IS NULL AND lead_id IS NOT NULL
    GROUP BY lead_id, business_id
) t ON t.lead_id = l.id AND t.business_id = l.business_id
SET l.next_follow_up = t.due, l.updated_at = l.updated_at
WHERE l.next_follow_up IS NULL;

UPDATE `crm_deals` d
JOIN (
    SELECT deal_id, business_id, MIN(due_date) AS due
    FROM `crm_tasks`
    WHERE completed_at IS NULL AND deal_id IS NOT NULL
    GROUP BY deal_id, business_id
) t ON t.deal_id = d.id AND t.business_id = d.business_id
SET d.next_follow_up = t.due, d.updated_at = d.updated_at
WHERE d.next_follow_up IS NULL;

-- Contracts close deals; one recorded without a lead is valid.
SET @contracts_missing := (
    SELECT COUNT(*) = 0 FROM information_schema.columns
    WHERE table_schema = DATABASE() AND table_name = 'crm_contracts' AND column_name = 'deal_id'
);
SET @contracts_sql := IF(@contracts_missing,
    'ALTER TABLE `crm_contracts`
       MODIFY `lead_id` bigint DEFAULT NULL,
       ADD COLUMN `deal_id` bigint DEFAULT NULL AFTER `lead_id`,
       ADD KEY `idx_crm_contracts_deal` (`business_id`,`deal_id`)',
    'SELECT 1');
PREPARE contracts_statement FROM @contracts_sql;
EXECUTE contracts_statement;
DEALLOCATE PREPARE contracts_statement;

UPDATE `crm_contracts` c
JOIN `crm_leads` l ON l.id = c.lead_id AND l.business_id = c.business_id
SET c.deal_id = l.deal_id, c.updated_at = c.updated_at
WHERE c.deal_id IS NULL AND l.deal_id IS NOT NULL;
