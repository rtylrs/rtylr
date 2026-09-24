-- name: GetSettings :one
SELECT owner_id, business_id, default_currency, fiscal_year_start, timezone,
       default_period, compare_previous, dashboard_widgets, favorite_reports
FROM insights_settings
WHERE business_id = ?;

-- name: UpsertSettings :exec
INSERT INTO insights_settings (
  owner_id, business_id, default_currency, fiscal_year_start, timezone,
  default_period, compare_previous, dashboard_widgets, favorite_reports
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
  default_currency = VALUES(default_currency),
  fiscal_year_start = VALUES(fiscal_year_start),
  timezone = VALUES(timezone),
  default_period = VALUES(default_period),
  compare_previous = VALUES(compare_previous),
  dashboard_widgets = VALUES(dashboard_widgets),
  favorite_reports = VALUES(favorite_reports),
  updated_at = CURRENT_TIMESTAMP;
