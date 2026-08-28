-- cross-schema read from hr_shifts and hr_staff — aggregate completed shift hours by employee and week.
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
GROUP BY s.staff_id, staff_name, week_start
ORDER BY week_start DESC, staff_name ASC;

-- cross-schema read from hr_shifts and hr_staff — aggregate the selected week's completed shift hours by employee.
-- name: GetTimesheetsForWeek :many
SELECT s.staff_id,
       CONCAT(st.first_name, ' ', st.last_name) AS staff_name,
       DATE(s.date - INTERVAL WEEKDAY(s.date) DAY) AS week_start,
       COALESCE(SUM(s.hours_worked), 0) AS total_hours,
       COALESCE(SUM(s.hours_worked) - SUM(s.overtime_hours), 0) AS regular_hours,
       COALESCE(SUM(s.overtime_hours), 0) AS overtime_hours
FROM hr_shifts s
JOIN hr_staff st ON st.id = s.staff_id AND st.business_id = s.business_id
WHERE s.business_id = ?
  AND s.status = 'completed'
  AND s.date >= ?
  AND s.date < ?
GROUP BY s.staff_id, staff_name, week_start
ORDER BY week_start DESC, staff_name ASC;
