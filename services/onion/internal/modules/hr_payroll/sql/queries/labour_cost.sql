-- name: GetProjectLabourCost :one
-- Compensation stays behind the HR boundary. Only aggregate cents and
-- coverage counts cross into consuming modules.
SELECT
    e.business_id,
    e.project_id,
    COALESCE(SUM(CASE
        WHEN emp.hourly_rate IS NOT NULL
        THEN ROUND(emp.hourly_rate * e.duration_minutes / 60 * 100)
        ELSE 0
    END), 0) AS cost_cents,
    COALESCE(SUM(CASE
        WHEN emp.hourly_rate IS NOT NULL THEN e.duration_minutes
        ELSE 0
    END), 0) AS costed_minutes,
    COALESCE(SUM(CASE
        WHEN emp.hourly_rate IS NULL THEN e.duration_minutes
        ELSE 0
    END), 0) AS uncosted_minutes,
    COUNT(DISTINCT CASE
        WHEN emp.hourly_rate IS NULL THEN e.employee_id
        ELSE NULL
    END) AS uncosted_employee_count
FROM hr_time_entries e
JOIN hr_employees emp
  ON emp.id = e.employee_id
 AND emp.business_id = ?
WHERE e.business_id = ?
  AND e.project_id = ?
  AND e.is_billable = TRUE
GROUP BY e.business_id, e.project_id;
