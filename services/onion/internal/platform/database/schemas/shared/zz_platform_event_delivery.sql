-- Event outbox and delivery queues backing tenant webhook delivery (onion#787)
-- and automation execution (onion#834).
--
-- One outbox, two consumers. Both features need the same machinery: something
-- durable to record that a domain action happened, and a queue per consumer
-- with retries that cannot double-send. Building them separately means building
-- it twice, so the tables are introduced together.
--
-- Shape follows shared/email_jobs.sql, which is the queue pattern already in
-- use by the worker: a status enum, an attempt counter against a cap, the last
-- error, and a due-time index the poller reads. Deliberate differences from
-- that table are noted where they occur.
--
-- Additive only. Three new tables, no column added to or removed from anything
-- that exists, and every statement is idempotent so a stopped migration can be
-- retried after the underlying error is corrected.

-- platform_events is the outbox. A producer writes a row inside the same
-- transaction as the domain write it describes, so an event is never published
-- for a change that rolled back, and a change is never made without its event.
CREATE TABLE IF NOT EXISTS platform_events (
    id            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,

    -- No DEFAULT 0, deliberately. Eleven flow_* tables carry
    -- `business_id BIGINT NOT NULL DEFAULT 0` and that default is what turns a
    -- forgotten tenant into a silent cross-tenant read (onion#691). This is a
    -- fan-out path touching every module's data, which makes it the worst place
    -- in the product to inherit that. A producer that does not know its tenant
    -- fails to insert instead.
    business_id   BIGINT NOT NULL,

    event_type    VARCHAR(64)  NOT NULL,
    subject_type  VARCHAR(64)  NOT NULL,
    subject_id    BIGINT       NOT NULL,
    payload       JSON         NOT NULL,

    -- Producer-supplied idempotency. Two attempts at the same domain action
    -- record one event, so a retried request cannot fan out twice.
    dedupe_key    VARCHAR(191) NOT NULL,

    occurred_at   TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    -- NULL until the dispatcher has expanded this event into delivery rows.
    dispatched_at TIMESTAMP    NULL,
    created_at    TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,

    UNIQUE KEY uk_platform_events_dedupe (business_id, dedupe_key),
    -- The dispatcher's read: undispatched rows in arrival order.
    KEY idx_platform_events_pending (dispatched_at, id),
    KEY idx_platform_events_business_type (business_id, event_type, occurred_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- One row per (registration, event). The unique key is the idempotency
-- guarantee: a dispatcher that runs twice over the same event cannot create a
-- second delivery, so a retry re-sends at most what the first attempt sent.
CREATE TABLE IF NOT EXISTS webhook_deliveries (
    id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    business_id     BIGINT NOT NULL,
    -- hub_webhooks.id and platform_events.id. No FOREIGN KEY: `shared` is the
    -- first directory in SCHEMA_MIGRATION_DIRS and hub_webhooks is created in
    -- `hub`, the last, so a constraint here cannot resolve at migration time.
    -- The dispatcher joins on these and skips a row whose registration has
    -- gone. Adding the constraints later is itself additive.
    webhook_id      BIGINT NOT NULL,
    event_id        BIGINT UNSIGNED NOT NULL,

    -- 'dead' is distinct from 'failed': failed is retryable and will be picked
    -- up again, dead has exhausted max_attempts and needs a human. Without the
    -- distinction a dead-lettered delivery is indistinguishable from one that
    -- is simply between attempts.
    status          ENUM('pending','delivering','delivered','failed','dead') NOT NULL DEFAULT 'pending',
    attempts        TINYINT UNSIGNED NOT NULL DEFAULT 0,
    max_attempts    TINYINT UNSIGNED NOT NULL DEFAULT 5,
    -- Backoff is expressed by moving this forward, so one tenant's dead
    -- endpoint waits its own backoff instead of holding the queue.
    next_attempt_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    response_status SMALLINT NULL,
    last_error      TEXT,
    delivered_at    TIMESTAMP NULL,
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY uk_webhook_deliveries_target (webhook_id, event_id),
    KEY idx_webhook_deliveries_due (status, next_attempt_at, id),
    KEY idx_webhook_deliveries_business (business_id, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- The automation side of the same shape. Separate table rather than a shared
-- one with a discriminator: the two consumers fail differently, retry
-- differently, and a webhook carries a response status an automation has no
-- equivalent for.
CREATE TABLE IF NOT EXISTS automation_runs (
    id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    business_id     BIGINT NOT NULL,
    -- crm_automations.id and platform_events.id. No FOREIGN KEY, same reason
    -- as above: crm is created after shared.
    automation_id   BIGINT NOT NULL,
    event_id        BIGINT UNSIGNED NOT NULL,

    status          ENUM('pending','running','succeeded','failed','dead') NOT NULL DEFAULT 'pending',
    attempts        TINYINT UNSIGNED NOT NULL DEFAULT 0,
    max_attempts    TINYINT UNSIGNED NOT NULL DEFAULT 3,
    next_attempt_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_error      TEXT,
    completed_at    TIMESTAMP NULL,
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    -- An automation fires at most once per event, which is what stops a retry
    -- of the dispatcher from awarding points or sending an email twice.
    UNIQUE KEY uk_automation_runs_target (automation_id, event_id),
    KEY idx_automation_runs_due (status, next_attempt_at, id),
    KEY idx_automation_runs_business (business_id, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
