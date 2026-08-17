-- name: ListDashboardWidgets :many
-- Dashboard widgets are a small, per-business layout config (a handful of
-- widget types per role), not a paginated list a caller ever pages through -
-- see hr_dashboard/internal/api/dashboard_service.go#GetDashboardWidgets. A
-- fixed hard cap (matching the products/dashboard.sql "LIMIT 101" precedent)
-- bounds the read without adding pagination scaffolding nothing consumes.
SELECT
    id, owner_id, business_id, role_id, widget_type,
    position_x, position_y, width, height,
    config, created_at, updated_at
FROM hr_dashboard_widgets
WHERE business_id = ?
ORDER BY position_y ASC, position_x ASC
LIMIT 500;
