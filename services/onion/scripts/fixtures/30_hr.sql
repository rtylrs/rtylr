-- HR fixtures: one department, two employees, payroll config, an
-- approved payroll run with two entries (covers payroll export, list,
-- and entry-fetch flows).
--
-- business_id is deliberately 1001 here, matching owner_id, and is NOT the
-- tenant id 1 used by businesses/users/pos_member_locations. HR's self-service
-- handlers pass the caller's own UserID down a parameter chain named ownerID,
-- and that value lands in the business_id filter at the SQL layer -- so these
-- rows have to carry the caller's user id, which the tests supply as
-- testutil.OwnerID (1001). Changing this to 1 or to testutil.BusinessID (2001)
-- makes every self-service payroll query miss, and makes the not-found test
-- pass for the wrong reason.

INSERT INTO hr_departments (id, owner_id, business_id, location_id, name, description)
VALUES
    (4001, 1001, 1001, 0, 'Operations', 'Floor + kitchen staff')
ON DUPLICATE KEY UPDATE name = VALUES(name), location_id = VALUES(location_id), business_id = VALUES(business_id);

INSERT INTO hr_employees (id, owner_id, business_id, department_id, first_name, last_name, email, employment_type, status, salary, hourly_rate, overtime_rate, start_date)
VALUES
    (4101, 1001, 1001, 4001, 'Alice', 'Worker', 'alice@test.local', 'full_time', 'active', 1200.00, NULL, NULL, '2025-01-01'),
    (4102, 1001, 1001, 4001, 'Bob',   'Hourly', 'bob@test.local',   'part_time', 'active', NULL,    8.00, 12.00, '2025-03-01')
ON DUPLICATE KEY UPDATE email = VALUES(email), business_id = VALUES(business_id);

INSERT INTO hr_staff (id, owner_id, business_id, department_id, first_name, last_name, email, role, employment_type, status, salary, hourly_rate, overtime_rate, start_date, apps)
VALUES
    (4101, 1001, 1001, 4001, 'Alice', 'Worker', 'alice@test.local', 'cashier', 'full_time', 'active', 1200.00, NULL, NULL,  '2025-01-01', '[]'),
    (4102, 1001, 1001, 4001, 'Bob',   'Hourly', 'bob@test.local',   'server',  'part_time', 'active', NULL,    8.00, 12.00, '2025-03-01', '[]')
ON DUPLICATE KEY UPDATE email = VALUES(email), apps = VALUES(apps), business_id = VALUES(business_id);

INSERT INTO hr_payroll_config (id, owner_id, business_id, period_type, pay_day_of_month, overtime_enabled)
VALUES
    (4201, 1001, 1001, 'monthly', 28, TRUE)
ON DUPLICATE KEY UPDATE period_type = VALUES(period_type), pay_day_of_month = VALUES(pay_day_of_month), business_id = VALUES(business_id);

INSERT INTO hr_payroll_runs (id, owner_id, business_id, period_start, period_end, status, created_by, total_gross, total_net)
VALUES
    (4301, 1001, 1001, '2026-03-01', '2026-03-31', 'approved', 1001, 2400.00, 2200.00)
ON DUPLICATE KEY UPDATE status = VALUES(status), business_id = VALUES(business_id);

INSERT INTO hr_payroll_entries (id, owner_id, business_id, payroll_run_id, employee_id, hours_worked, overtime_hours, base_pay, overtime_pay, bonuses, deductions, net_pay)
VALUES
    (4401, 1001, 1001, 4301, 4101, 160.00, 0.00,  1200.00, 0.00,   0.00, 100.00, 1100.00),
    (4402, 1001, 1001, 4301, 4102,  80.00, 5.00,   640.00, 60.00,  0.00, 100.00, 600.00)
ON DUPLICATE KEY UPDATE net_pay = VALUES(net_pay), business_id = VALUES(business_id);
