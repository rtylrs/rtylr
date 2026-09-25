-- Feature support: flag definitions with percentage rollout + per-user targets.
-- These are GLOBAL (platform-level), not tenant-scoped: a flag controls a
-- feature/variant across all users. Per-user targeting lives in feature_targets.
-- Designed to grow: add feature_segments / conditions tables later for richer
-- targeting without changing this core.

CREATE TABLE IF NOT EXISTS feature_flags (
  id                 BIGINT       NOT NULL AUTO_INCREMENT,
  feature_key        VARCHAR(120) NOT NULL,
  description        VARCHAR(255) NOT NULL DEFAULT '',
  -- master switch: when 0, the feature is off for everyone (variant_off)
  enabled            TINYINT(1)   NOT NULL DEFAULT 0,
  -- 0..100 — the share of subjects deterministically bucketed into variant_on
  rollout_percentage INT          NOT NULL DEFAULT 0,
  -- variant returned when a subject is bucketed IN / a target forces ON
  variant_on         VARCHAR(60)  NOT NULL DEFAULT 'b',
  -- variant returned when bucketed OUT / disabled (the control / default)
  variant_off        VARCHAR(60)  NOT NULL DEFAULT 'a',
  created_at         TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at         TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uq_feature_flags_key (feature_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS feature_targets (
  id          BIGINT       NOT NULL AUTO_INCREMENT,
  feature_key VARCHAR(120) NOT NULL,
  -- the targeted user (== users.id == owner_id elsewhere in the system)
  owner_id    BIGINT       NOT NULL,
  -- force the feature on/off for this specific user, overriding the rollout
  enabled     TINYINT(1)   NOT NULL DEFAULT 1,
  -- optional explicit variant for this user; NULL -> use the flag's variant_on/off
  variant     VARCHAR(60)  NULL,
  created_at  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uq_feature_targets (feature_key, owner_id),
  KEY idx_feature_targets_owner (owner_id),
  KEY idx_feature_targets_key (feature_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- feature_rules: ordered targeting rules per flag. A rule matches when ALL of
-- its conditions match the evaluation context (the attributes about the
-- subject). The first matching rule (lowest priority number) decides the
-- outcome — serving `variant` to `rollout_percentage`% of the matched audience
-- (the rest get the flag's variant_off). This is how you "target by anything":
-- role, plan, app, platform, country, email, custom attributes, etc.
CREATE TABLE IF NOT EXISTS feature_rules (
  id                 BIGINT       NOT NULL AUTO_INCREMENT,
  feature_key        VARCHAR(120) NOT NULL,
  priority           INT          NOT NULL DEFAULT 0,
  description        VARCHAR(255) NOT NULL DEFAULT '',
  -- variant served to subjects this rule matches + buckets in
  variant            VARCHAR(60)  NOT NULL DEFAULT 'b',
  -- % of the MATCHED audience that gets `variant` (rest get the flag's off)
  rollout_percentage INT          NOT NULL DEFAULT 100,
  enabled            TINYINT(1)   NOT NULL DEFAULT 1,
  created_at         TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at         TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_feature_rules_key_priority (feature_key, priority)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- feature_rule_conditions: the predicates of a rule. attribute is a key into
-- the evaluation context (e.g. 'role', 'plan_id', 'country', 'app', 'email',
-- 'business_id', 'user_id', or any custom 'beta'), operator is how to compare,
-- value is the comparison operand (comma-separated for in/nin).
CREATE TABLE IF NOT EXISTS feature_rule_conditions (
  id         BIGINT       NOT NULL AUTO_INCREMENT,
  rule_id    BIGINT       NOT NULL,
  attribute  VARCHAR(120) NOT NULL,
  -- eq | neq | in | nin | contains | prefix | suffix | gt | gte | lt | lte | regex | exists
  operator   VARCHAR(20)  NOT NULL DEFAULT 'eq',
  value      TEXT         NOT NULL,
  created_at TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_feature_rule_conditions_rule (rule_id),
  CONSTRAINT fk_feature_rule_conditions_rule
    FOREIGN KEY (rule_id) REFERENCES feature_rules (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO feature_flags (
  feature_key,
  description,
  enabled,
  rollout_percentage,
  variant_on,
  variant_off
)
VALUES (
  'is_desktop_mode_default',
  'Default the rtylr landing page to the desktop OS experience when enabled.',
  FALSE,
  0,
  'true',
  'false'
)
ON DUPLICATE KEY UPDATE
  description = VALUES(description),
  enabled = VALUES(enabled),
  rollout_percentage = VALUES(rollout_percentage),
  variant_on = VALUES(variant_on),
  variant_off = VALUES(variant_off);
