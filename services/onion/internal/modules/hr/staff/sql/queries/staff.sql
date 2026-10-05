-- name: GetStaffByID :one
SELECT id, owner_id, business_id, user_id, department_id, manager_staff_id, first_name, last_name, email, phone,
       role, employment_type, status, salary, hourly_rate, overtime_rate,
       start_date, end_date, address, emergency_contact_name, emergency_contact_phone,
       apps, notes, photo_url, bio, created_at, updated_at
FROM hr_staff
WHERE id = ? AND business_id = ?;

-- name: CreateStaff :execresult
INSERT INTO hr_staff
(owner_id, business_id, user_id, department_id, manager_staff_id, first_name, last_name, email, phone, role,
 employment_type, status, salary, hourly_rate, overtime_rate, start_date, end_date,
 address, emergency_contact_name, emergency_contact_phone, apps, notes, photo_url, bio)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateStaff :exec
UPDATE hr_staff
SET department_id = ?, manager_staff_id = ?, first_name = ?, last_name = ?, email = ?, phone = ?,
    role = ?, employment_type = ?, status = ?, salary = ?, hourly_rate = ?,
    overtime_rate = ?, end_date = ?, address = ?, emergency_contact_name = ?,
    emergency_contact_phone = ?, apps = ?, notes = ?, photo_url = ?, bio = ?,
    updated_at = NOW()
WHERE id = ? AND business_id = ?;

-- name: TerminateStaff :exec
UPDATE hr_staff SET status = 'terminated', updated_at = NOW() WHERE id = ? AND business_id = ?;

-- name: CountActiveStaff :one
SELECT COUNT(*) FROM hr_staff WHERE business_id = ? AND status = 'active';

-- cross-schema read from hr_shifts — show staff currently assigned to a shift.
-- name: CountOnShiftToday :one
SELECT COUNT(*) FROM hr_shifts WHERE business_id = ? AND date = CURDATE() AND status IN ('scheduled','in_progress');

-- cross-schema read from hr_time_off_requests — show pending time-off requests in staff dashboard counts.
-- name: CountPendingTimeOff :one
SELECT COUNT(*) FROM hr_time_off_requests WHERE business_id = ? AND status = 'pending';

-- cross-schema read from hr_payroll_records — show draft payroll records in staff dashboard counts.
-- name: CountDraftPayroll :one
SELECT COUNT(*) FROM hr_payroll_records WHERE business_id = ? AND status = 'draft';

-- name: ClockIn :exec
UPDATE hr_shifts SET clock_in_at = NOW(), status = 'in_progress', updated_at = NOW()
WHERE business_id = ? AND staff_id = ? AND date = CURDATE() AND clock_in_at IS NULL
ORDER BY start_time ASC LIMIT 1;

-- name: ClockOut :exec
UPDATE hr_shifts SET clock_out_at = NOW(), status = 'completed',
       hours_worked = TIMESTAMPDIFF(MINUTE, clock_in_at, NOW()) / 60.0, updated_at = NOW()
WHERE business_id = ? AND staff_id = ? AND date = CURDATE() AND clock_in_at IS NOT NULL AND clock_out_at IS NULL
ORDER BY start_time ASC LIMIT 1;

-- cross-schema read from hr_departments — resolve department-managed reports for Flow visibility.
-- name: ListManagedUserIDs :many
-- Users reporting to a manager, for the Flow dept-manager visibility proxy (M3).
-- The relationship lives on hr_departments.manager_staff_id, so this joins back
-- to hr_staff to identify the manager by their USER id, which is what Flow has.
--
-- Direct departments only: hr_departments.parent_department_id exists, but a
-- manager of a parent department does not inherit the children here. Making
-- visibility transitive up an org tree is a larger decision than this proxy,
-- and a recursive walk would silently widen access as the tree grows.
--
-- Terminated staff are excluded: they are no longer a report. Staff with no
-- linked user account are excluded because they cannot be a Flow subject.
-- Two rule sources, unioned. The department manager (below) and, since F-B, a
-- direct reporting line on the staff record. orgchart.ManagerResolver's contract
-- says consumers must not care which rule produced an id, so adding the second
-- is not a breaking change.
SELECT DISTINCT s.user_id
FROM hr_staff s
JOIN hr_departments d ON d.id = s.department_id AND d.business_id = s.business_id
JOIN hr_staff m ON m.id = d.manager_staff_id AND m.business_id = s.business_id
WHERE m.user_id = ?
  AND s.business_id = ?
  AND s.user_id IS NOT NULL
  AND s.user_id <> m.user_id
  AND s.status <> 'terminated'
UNION
-- Direct reports: staff whose manager_staff_id points at the caller's own staff
-- row. s.manager_staff_id <> s.id excludes a record that names itself, which is
-- why the schema carries no CHECK for it.
SELECT DISTINCT s2.user_id
FROM hr_staff s2
JOIN hr_staff m2 ON m2.id = s2.manager_staff_id AND m2.business_id = s2.business_id
WHERE m2.user_id = ?
  AND s2.business_id = ?
  AND s2.user_id IS NOT NULL
  AND s2.user_id <> m2.user_id
  AND s2.manager_staff_id <> s2.id
  AND s2.status <> 'terminated';
