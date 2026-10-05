-- Dedicated rows for operation-level authorization evidence. These records
-- are isolated from the ordinary integration fixtures so direct tests can
-- exercise reads and mutations without depending on test ordering.

UPDATE hr_employees
SET user_id = CASE id WHEN 4101 THEN 1001 WHEN 4102 THEN 1002 END,
    business_id = 1001
WHERE id IN (4101, 4102);

INSERT INTO careers
    (id, business_id, owner_id, title, slug, details, is_active, is_published)
VALUES
    (7001, 1001, 1001, 'Direct Authorization Role', 'direct-authorization-role', 'Role used by direct authorization evidence.', TRUE, FALSE)
ON DUPLICATE KEY UPDATE business_id = VALUES(business_id), owner_id = VALUES(owner_id), is_active = VALUES(is_active);

INSERT INTO careers_submissions
    (id, career_id, owner_id, business_id, candidate_name, email, phone, resume_url, cover_letter, status, answers_json)
VALUES
    (7002, 7001, 1001, 1001, 'Direct Candidate', 'candidate@test.local', '+961-3-700200', 'https://storage.test.local/direct-resume.pdf', 'Direct authorization submission.', 'new', JSON_OBJECT('source', 'direct'))
ON DUPLICATE KEY UPDATE career_id = VALUES(career_id), owner_id = VALUES(owner_id), business_id = VALUES(business_id), resume_url = VALUES(resume_url), status = VALUES(status);

INSERT INTO crm_leads
    (id, owner_id, business_id, company_name, industry, size, source, stage, assigned_to, notes)
VALUES
    (7101, 1001, 1001, 'Direct Authorization Lead', 'software', 'small', 'test', 'cold', 1001, 'Direct authorization fixture.')
ON DUPLICATE KEY UPDATE owner_id = VALUES(owner_id), business_id = VALUES(business_id), assigned_to = VALUES(assigned_to);

INSERT INTO crm_contacts
    (id, owner_id, business_id, lead_id, name, role, email, phone, is_primary, notes)
VALUES
    (7102, 1001, 1001, 7101, 'Direct Contact', 'Owner', 'contact@test.local', '+961-3-710200', TRUE, 'Direct authorization fixture.')
ON DUPLICATE KEY UPDATE owner_id = VALUES(owner_id), business_id = VALUES(business_id), lead_id = VALUES(lead_id), name = VALUES(name);

INSERT INTO crm_campaigns
    (id, owner_id, business_id, name, description, channel, status, subject, content)
VALUES
    (7006, 1001, 1001, 'Direct Authorization Campaign', 'Direct authorization fixture campaign.', 'email', 'draft', 'Direct authorization', 'Direct authorization campaign content.')
ON DUPLICATE KEY UPDATE business_id = VALUES(business_id), status = VALUES(status), subject = VALUES(subject), content = VALUES(content);

INSERT INTO hr_roles
    (id, owner_id, business_id, name, description, permissions, is_system)
VALUES
    (7201, 1001, 1001, 'Direct Authorization HR Operator', 'Direct authorization fixture role.', JSON_OBJECT('approve_time_off', TRUE, 'manage_payroll', TRUE), FALSE)
ON DUPLICATE KEY UPDATE business_id = VALUES(business_id), permissions = VALUES(permissions);

INSERT INTO hr_employee_roles
    (id, owner_id, business_id, employee_id, role_id)
VALUES
    (7202, 1001, 1001, 4101, 7201)
ON DUPLICATE KEY UPDATE business_id = VALUES(business_id);

INSERT INTO hr_leave_types
    (id, owner_id, business_id, name, is_paid, requires_approval, min_notice_days, is_active)
VALUES
    (7203, 1001, 1001, 'Direct Authorization Leave', TRUE, FALSE, 0, TRUE)
ON DUPLICATE KEY UPDATE business_id = VALUES(business_id), requires_approval = VALUES(requires_approval), is_active = VALUES(is_active);

INSERT INTO hr_leave_balances
    (id, owner_id, business_id, employee_id, leave_type_id, year, entitled_days, used_days, pending_days)
VALUES
    (7220, 1001, 1001, 4101, 7203, 2026, 30.00, 0.00, 0.00),
    (7204, 1001, 1001, 4102, 7203, 2026, 30.00, 0.00, 0.00)
ON DUPLICATE KEY UPDATE business_id = VALUES(business_id), employee_id = VALUES(employee_id), entitled_days = VALUES(entitled_days), used_days = VALUES(used_days), pending_days = VALUES(pending_days);

INSERT INTO hr_leave_requests
    (id, owner_id, business_id, employee_id, leave_type_id, start_date, end_date, total_days, notes, status, current_approval_level)
VALUES
    (7205, 1001, 1001, 4102, 7203, '2026-12-10', '2026-12-11', 2.00, 'Direct authorization fixture.', 'pending', 0)
ON DUPLICATE KEY UPDATE business_id = VALUES(business_id), employee_id = VALUES(employee_id), status = VALUES(status), current_approval_level = VALUES(current_approval_level);

INSERT INTO insights_scheduled_reports
    (id, owner_id, business_id, template_id, cadence, channel, recipient, is_active)
VALUES
    (7206, 1001, 1001, 'revenue-monthly', 'monthly', 'email', 'owner@test.local', TRUE)
ON DUPLICATE KEY UPDATE business_id = VALUES(business_id), template_id = VALUES(template_id), is_active = VALUES(is_active);

INSERT INTO pos_orders
    (id, owner_id, business_id, session_id, table_id, location_id, order_number, type, status, payment_status, customer_name, staff_id, subtotal, total, amount_paid)
VALUES
    (7211, 1001, 1001, 3201, 3101, 3001, 'AUTH-VOID', 'dine_in', 'pending', 'unpaid', 'Direct Authorization', 1002, 20.00, 20.00, 0.00),
    (7212, 1001, 1001, 3201, 3101, 3001, 'AUTH-VOID-APPROVE', 'dine_in', 'pending', 'unpaid', 'Direct Authorization', 1002, 21.00, 21.00, 0.00),
    (7213, 1001, 1001, 3201, 3101, 3001, 'AUTH-REFUND', 'dine_in', 'paid', 'paid', 'Direct Authorization', 1002, 22.00, 22.00, 22.00),
    (7214, 1001, 1001, 3201, 3101, 3001, 'AUTH-REFUND-APPROVE', 'dine_in', 'paid', 'paid', 'Direct Authorization', 1002, 23.00, 23.00, 23.00),
    (7215, 1001, 1001, 3201, 3101, 3001, 'AUTH-VOID-OVERRIDE', 'dine_in', 'pending', 'unpaid', 'Direct Authorization', 1002, 24.00, 24.00, 0.00),
    (7216, 1001, 1001, 3201, 3101, 3001, 'AUTH-REFUND-OVERRIDE', 'dine_in', 'paid', 'paid', 'Direct Authorization', 1002, 25.00, 25.00, 25.00),
    (7217, 1001, 1001, 3201, 3101, 3001, 'AUTH-DISCOUNT', 'dine_in', 'paid', 'paid', 'Direct Authorization', 1002, 26.00, 26.00, 26.00),
    (7218, 1001, 1001, 3201, 3101, 3001, 'AUTH-DISCOUNT-STAFF', 'dine_in', 'pending', 'unpaid', 'Direct Authorization', 1001, 26.00, 26.00, 0.00)
ON DUPLICATE KEY UPDATE business_id = VALUES(business_id), status = VALUES(status), payment_status = VALUES(payment_status), total = VALUES(total), amount_paid = VALUES(amount_paid);

INSERT INTO pos_sessions
    (id, owner_id, business_id, location_id, staff_id, status, opening_cash, total_sales, total_orders)
VALUES
    (7210, 1001, 1001, 3001, 1001, 'open', 100.00, 0.00, 0)
ON DUPLICATE KEY UPDATE business_id = VALUES(business_id), staff_id = VALUES(staff_id), status = VALUES(status);

INSERT INTO erp_purchase_orders
    (id, owner_id, business_id, supplier_id, po_number, status, subtotal, total, created_by)
VALUES
    (7301, 1001, 1001, 5101, 'AUTH-APPROVE', 'draft', 5.00, 5.00, 1002)
ON DUPLICATE KEY UPDATE business_id = VALUES(business_id), status = VALUES(status), created_by = VALUES(created_by);

INSERT INTO erp_purchase_order_items
    (id, purchase_order_id, product_id, quantity_ordered, quantity_received, unit_cost, total_cost)
VALUES
    (7302, 7301, 5201, 1.000, 0.000, 5.00, 5.00)
ON DUPLICATE KEY UPDATE product_id = VALUES(product_id), total_cost = VALUES(total_cost);

INSERT INTO erp_stock_takes
    (id, owner_id, business_id, warehouse_id, status, notes)
VALUES
    (7303, 1001, 1001, 1, 'draft', 'Direct authorization fixture.')
ON DUPLICATE KEY UPDATE business_id = VALUES(business_id), warehouse_id = VALUES(warehouse_id), status = VALUES(status);

INSERT INTO erp_stock_take_items
    (id, stock_take_id, product_id, expected_qty, counted_qty, variance)
VALUES
    (7304, 7303, 5201, 50.000, NULL, NULL)
ON DUPLICATE KEY UPDATE product_id = VALUES(product_id), expected_qty = VALUES(expected_qty), counted_qty = VALUES(counted_qty), variance = VALUES(variance);

INSERT INTO pos_terminals
    (id, business_id, branch_id, name, device_id, is_active)
VALUES
    (1, 1001, 3001, 'Direct Authorization Terminal', 'direct-authorization-terminal', TRUE)
ON DUPLICATE KEY UPDATE business_id = VALUES(business_id), branch_id = VALUES(branch_id), is_active = VALUES(is_active);

INSERT INTO pos_users
    (id, owner_id, business_id, user_id, location_id, name, role, is_active, max_discount_pct, can_apply_discount)
VALUES
    (7219, 1001, 1001, 1001, 3001, 'Direct Authorization Operator', 'manager', TRUE, 100.00, TRUE)
ON DUPLICATE KEY UPDATE business_id = VALUES(business_id), user_id = VALUES(user_id), location_id = VALUES(location_id), is_active = VALUES(is_active), max_discount_pct = VALUES(max_discount_pct), can_apply_discount = VALUES(can_apply_discount);

INSERT INTO authorization_roles
    (id, business_id, app_id, role_key, display_name, description, level, assignment_mode, is_system, is_archived, lifecycle, version, created_by)
VALUES
    (7305, 1001, 'dash', 'direct-authorization-role', 'Direct Authorization Role', 'Direct authorization fixture role.', 10, 'many', FALSE, FALSE, 'active', 1, 1001),
    (7308, 1001, 'dash', 'direct-team-authorization-role', 'Direct Team Authorization Role', 'Isolated profile for team authorization evidence.', 10, 'many', FALSE, FALSE, 'active', 1, 1001)
ON DUPLICATE KEY UPDATE
    business_id = VALUES(business_id), display_name = VALUES(display_name),
    is_archived = VALUES(is_archived), lifecycle = VALUES(lifecycle);

INSERT INTO authorization_roles
    (id, business_id, app_id, role_key, display_name, description, level, assignment_mode, is_system, is_archived, lifecycle, version, created_by)
VALUES
    (7306, 1001, 'pos', 'direct-authorization-pos-role', 'Direct Authorization POS Role', 'POS profile used by invalid-scope tests.', 10, 'many', FALSE, FALSE, 'active', 1, 1001)
ON DUPLICATE KEY UPDATE business_id = VALUES(business_id), display_name = VALUES(display_name), is_archived = VALUES(is_archived), lifecycle = VALUES(lifecycle);

INSERT INTO finance_invoices
    (id, owner_id, business_id, invoice_number, client_name, status, issue_date, due_date, currency, subtotal, total, amount_due)
VALUES
    (7401, 1001, 1001, 'AUTH-SEND', 'Direct Authorization', 'draft', '2026-08-27', '2026-09-27', 'USD', 1.00, 1.00, 1.00)
ON DUPLICATE KEY UPDATE business_id = VALUES(business_id), status = VALUES(status), total = VALUES(total), amount_due = VALUES(amount_due);

INSERT INTO authorization_business_members
    (business_id, user_id, member_type, status)
VALUES
    (1001, 1002, 'member', 'active'),
    (1001, 1003, 'member', 'active')
ON DUPLICATE KEY UPDATE member_type = VALUES(member_type), status = VALUES(status);

INSERT INTO user_provisioning_states
    (user_id, business_id, mode, state, email_verified, password_change_required, mfa_required, mfa_enrolled)
VALUES
    (1001, 1001, 'invitation', 'active', TRUE, FALSE, FALSE, FALSE),
    (1002, 1001, 'invitation', 'active', TRUE, FALSE, FALSE, FALSE),
    (1003, 1001, 'invitation', 'active', TRUE, FALSE, FALSE, FALSE)
ON DUPLICATE KEY UPDATE
    business_id = VALUES(business_id), state = VALUES(state), email_verified = VALUES(email_verified),
    password_change_required = VALUES(password_change_required), mfa_required = VALUES(mfa_required),
    mfa_enrolled = VALUES(mfa_enrolled);
