-- name: GetPOSSettings :one
SELECT id, business_id, business_profile, theme, features, receipt_footer, created_at, updated_at
FROM pos_business_settings
WHERE business_id = ?
LIMIT 1;

-- name: UpsertPOSSettings :exec
INSERT INTO pos_business_settings (business_id, business_profile, theme, features, receipt_footer)
VALUES (?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
  business_profile = VALUES(business_profile),
  theme = VALUES(theme),
  features = VALUES(features),
  receipt_footer = VALUES(receipt_footer),
  updated_at = CURRENT_TIMESTAMP;
