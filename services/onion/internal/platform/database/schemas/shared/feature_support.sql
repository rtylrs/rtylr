-- Platform feature gates. This mirrors the featuresupport module schema so the
-- normal database migration path actually provisions it.

CREATE TABLE IF NOT EXISTS feature_flags (
  id                 BIGINT       NOT NULL AUTO_INCREMENT,
  feature_key        VARCHAR(120) NOT NULL,
  description        VARCHAR(255) NOT NULL DEFAULT '',
  enabled            TINYINT(1)   NOT NULL DEFAULT 0,
  rollout_percentage INT          NOT NULL DEFAULT 0,
  variant_on         VARCHAR(60)  NOT NULL DEFAULT 'b',
  variant_off        VARCHAR(60)  NOT NULL DEFAULT 'a',
  created_at         TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at         TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uq_feature_flags_key (feature_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS feature_targets (
  id          BIGINT       NOT NULL AUTO_INCREMENT,
  feature_key VARCHAR(120) NOT NULL,
  owner_id    BIGINT       NOT NULL,
  enabled     TINYINT(1)   NOT NULL DEFAULT 1,
  variant     VARCHAR(60)  NULL,
  created_at  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uq_feature_targets (feature_key, owner_id),
  KEY idx_feature_targets_owner (owner_id),
  KEY idx_feature_targets_key (feature_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS feature_rules (
  id                 BIGINT       NOT NULL AUTO_INCREMENT,
  feature_key        VARCHAR(120) NOT NULL,
  priority           INT          NOT NULL DEFAULT 0,
  description        VARCHAR(255) NOT NULL DEFAULT '',
  variant            VARCHAR(60)  NOT NULL DEFAULT 'b',
  rollout_percentage INT          NOT NULL DEFAULT 100,
  enabled            TINYINT(1)   NOT NULL DEFAULT 1,
  created_at         TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at         TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_feature_rules_key_priority (feature_key, priority)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS feature_rule_conditions (
  id         BIGINT       NOT NULL AUTO_INCREMENT,
  rule_id    BIGINT       NOT NULL,
  attribute  VARCHAR(120) NOT NULL,
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
  'auth.social_login_enabled',
  'Show social-login buttons on the auth login screen.',
  TRUE,
  100,
  'on',
  'off'
)
ON DUPLICATE KEY UPDATE
  feature_key = feature_key;

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
