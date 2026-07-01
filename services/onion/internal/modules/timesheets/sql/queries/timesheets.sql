-- name: GetTimesheets :many
SELECT s.staff_id,
       CONCAT(st.first_name, ' ', st.last_name) AS staff_name,
       DATE(s.date - INTERVAL WEEKDAY(s.date) DAY) AS week_start,
       COALESCE(SUM(s.hours_worked), 0) AS total_hours,
       COALESCE(SUM(s.hours_worked) - SUM(s.overtime_hours), 0) AS regular_hours,
       COALESCE(SUM(s.overtime_hours), 0) AS overtime_hours
FROM hr_shifts s
JOIN hr_staff st ON st.id = s.staff_id AND st.business_id = s.business_id
WHERE s.business_id = ? AND s.status = 'completed'
  AND (sqlc.narg('week_start') IS NULL OR DATE(s.date - INTERVAL WEEKDAY(s.date) DAY) = sqlc.narg('week_start'))
GROUP BY s.staff_id, staff_name, week_start
ORDER BY week_start DESC, staff_name ASC;

