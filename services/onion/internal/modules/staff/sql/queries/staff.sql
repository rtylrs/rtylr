-- name: GetStaffByID :one
SELECT id, owner_id, business_id, user_id, department_id, first_name, last_name, email, phone,
       role, employment_type, status, salary, hourly_rate, overtime_rate,
       start_date, end_date, address, emergency_contact_name, emergency_contact_phone,
       apps, notes, photo_url, bio, created_at, updated_at
FROM hr_staff
WHERE id = ? AND business_id = ?;

-- name: CreateStaff :execresult
INSERT INTO hr_staff
(owner_id, business_id, user_id, department_id, first_name, last_name, email, phone, role,
 employment_type, status, salary, hourly_rate, overtime_rate, start_date, end_date,
 address, emergency_contact_name, emergency_contact_phone, apps, notes, photo_url, bio)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateStaff :exec
UPDATE hr_staff
SET department_id = ?, first_name = ?, last_name = ?, email = ?, phone = ?,
    role = ?, employment_type = ?, status = ?, salary = ?, hourly_rate = ?,
    overtime_rate = ?, end_date = ?, address = ?, emergency_contact_name = ?,
    emergency_contact_phone = ?, apps = ?, notes = ?, photo_url = ?, bio = ?,
    updated_at = NOW()
WHERE id = ? AND business_id = ?;

-- name: TerminateStaff :exec
UPDATE hr_staff SET status = 'terminated', updated_at = NOW() WHERE id = ? AND business_id = ?;

-- name: CountActiveStaff :one
SELECT COUNT(*) FROM hr_staff WHERE business_id = ? AND status = 'active';

-- name: CountOnShiftToday :one
SELECT COUNT(*) FROM hr_shifts WHERE business_id = ? AND date = CURDATE() AND status IN ('scheduled','in_progress');

-- name: CountPendingTimeOff :one
SELECT COUNT(*) FROM hr_time_off_requests WHERE business_id = ? AND status = 'pending';

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
