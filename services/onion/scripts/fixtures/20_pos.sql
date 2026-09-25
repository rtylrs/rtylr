-- POS fixtures: a single location, two tables, an open session, two
-- orders (one paid + one refunded), and a couple of POS staff users —
-- enough to exercise list / refund / session-totals queries.

INSERT INTO pos_locations (id, owner_id, business_id, name, address, phone, is_active)
VALUES
    (3001, 1001, 1001, 'Test Branch', '123 Test St', '+961-1-000000', TRUE)
ON DUPLICATE KEY UPDATE name = VALUES(name);

INSERT INTO pos_tables (id, owner_id, business_id, location_id, name, capacity, status)
VALUES
    (3101, 1001, 1001, 3001, 'T1', 4, 'available'),
    (3102, 1001, 1001, 3001, 'T2', 2, 'occupied')
ON DUPLICATE KEY UPDATE name = VALUES(name);

INSERT INTO pos_sessions (id, owner_id, business_id, location_id, staff_id, status, opening_cash, total_sales, total_orders)
VALUES
    (3201, 1001, 1001, 3001, 3601, 'open', 100.00, 250.00, 2)
ON DUPLICATE KEY UPDATE status = VALUES(status);

INSERT INTO pos_orders (id, owner_id, business_id, session_id, table_id, location_id, order_number, type, status, payment_status, customer_name, staff_id, subtotal, total)
VALUES
    (3301, 1001, 1001, 3201, 3101, 3001, 'ORD-1', 'dine_in', 'paid',     'paid',    'Walk-in', 1002, 150.00, 150.00),
    (3302, 1001, 1001, 3201, 3102, 3001, 'ORD-2', 'dine_in', 'refunded', 'refunded','Walk-in', 1002, 100.00, 100.00)
ON DUPLICATE KEY UPDATE status = VALUES(status), payment_status = VALUES(payment_status);

INSERT INTO pos_order_items (id, order_id, menu_item_id, name, quantity, unit_price, total)
VALUES
    (3401, 3301, 5201, 'Burger',  2, 50.00, 100.00),
    (3402, 3301, 5202, 'Soda',    2, 25.00,  50.00),
    (3403, 3302, 5203, 'Salad',   1, 60.00,  60.00),
    (3404, 3302, 5204, 'Coffee',  2, 20.00,  40.00)
ON DUPLICATE KEY UPDATE menu_item_id = VALUES(menu_item_id), total = VALUES(total);

INSERT INTO pos_refunds (id, owner_id, order_id, amount, reason, refunded_by)
VALUES
    (3501, 1001, 3302, 100.00, 'Customer request', 1002)
ON DUPLICATE KEY UPDATE amount = VALUES(amount);

-- POS staff are team users with a per-location PIN credential. user_id links
-- each pos_users row back to the users table; pos_member_locations records the
-- location assignment that authorises the PIN.
INSERT INTO pos_member_locations (id, business_id, user_id, location_id, can_manage_pos_users, can_manage_terminals, can_view_reports)
VALUES
    (3701, 1001, 1002, 3001, FALSE, FALSE, FALSE),
    (3702, 1001, 1003, 3001, FALSE, FALSE, FALSE)
ON DUPLICATE KEY UPDATE location_id = VALUES(location_id);

-- pin_hash, not the dropped plaintext pin: POS PINs are bcrypt hashes
-- (zz_pos_users_remove_plaintext_pin.sql removed the legacy column).
INSERT INTO pos_users (id, owner_id, business_id, user_id, location_id, name, role, is_active)
VALUES
    (3601, 1001, 1001, 1002, 3001, 'Cashier A', 'cashier', TRUE),
    (3602, 1001, 1001, 1003, 3001, 'Server B',  'server',  TRUE)
ON DUPLICATE KEY UPDATE name = VALUES(name), business_id = VALUES(business_id);
