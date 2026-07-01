-- name: ListDashboardWidgets :many
SELECT
    id, owner_id, business_id, role_id, widget_type,
    position_x, position_y, width, height,
    config, created_at, updated_at
FROM hr_dashboard_widgets
WHERE business_id = ?
ORDER BY position_y ASC, position_x ASC;
