-- name: GetSettings :one
SELECT owner_id, business_id, default_currency, fiscal_year_start, timezone
FROM insights_settings
WHERE business_id = ?;

-- name: UpsertSettings :exec
INSERT INTO insights_settings (owner_id, business_id, default_currency, fiscal_year_start, timezone)
VALUES (?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
  default_currency = VALUES(default_currency),
  fiscal_year_start = VALUES(fiscal_year_start),
  timezone = VALUES(timezone),
  updated_at = CURRENT_TIMESTAMP;
