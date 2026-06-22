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
