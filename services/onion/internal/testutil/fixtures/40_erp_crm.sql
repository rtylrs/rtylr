-- ERP and CRM fixtures: a couple of products with stock, a supplier, a
-- category, and two customers (one bronze, one gold) so list and search
-- queries return rows.

INSERT INTO erp_categories (id, owner_id, business_id, name, description)
VALUES
    (5001, 1001, 1001, 'Beverages', 'Drinks'),
    (5002, 1001, 1001, 'Mains',     'Main dishes')
ON DUPLICATE KEY UPDATE name = VALUES(name);

INSERT INTO erp_suppliers (id, owner_id, business_id, name, email, phone, is_active)
VALUES
    (5101, 1001, 1001, 'Test Supplier', 'supplier@test.local', '+961-1-111111', TRUE)
ON DUPLICATE KEY UPDATE name = VALUES(name);

INSERT INTO erp_products (id, owner_id, business_id, category_id, supplier_id, name, sku, cost_price, sale_price, stock_quantity, low_stock_threshold, is_active, track_inventory)
VALUES
    (5201, 1001, 1001, 5002, 5101, 'Burger', 'BURG-1', 5.00,  10.00, 50.00, 10.00, TRUE, TRUE),
    (5202, 1001, 1001, 5001, 5101, 'Soda',   'SODA-1', 0.50,   2.50, 200.00, 50.00, TRUE, TRUE),
    (5203, 1001, 1001, 5002, 5101, 'Salad',  'SAL-1',  3.00,   8.00,  4.00, 10.00, TRUE, TRUE),
    (5204, 1001, 1001, 5001, 5101, 'Coffee', 'COF-1',  1.00,   3.50, 75.00, 20.00, TRUE, TRUE)
ON DUPLICATE KEY UPDATE name = VALUES(name);

INSERT INTO crm_customers (id, owner_id, business_id, first_name, last_name, email, phone, loyalty_points, total_spent, visit_count, loyalty_tier, tags, consent)
VALUES
    (5301, 1001, 1001, 'Jamie', 'Bronze', 'jamie@test.local', '+961-3-100000', 50,  100.00, 5,  'bronze', '[]', JSON_OBJECT()),
    (5302, 1001, 1001, 'Alex',  'Gold',   'alex@test.local',  '+961-3-200000', 1500, 2500.00, 30, 'gold',   '[]', JSON_OBJECT())
ON DUPLICATE KEY UPDATE email = VALUES(email), tags = VALUES(tags);
