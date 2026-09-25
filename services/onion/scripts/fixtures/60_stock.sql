-- Shared stock fixtures mirror the ERP product fixture so inventory reports,
-- valuation, and stock lifecycle integration tests have a real location.
INSERT INTO erp_warehouses
    (id, owner_id, business_id, name, address, city, country, manager_name, capacity, is_default, is_active)
VALUES
    (1, 1001, 1001, 'Main Warehouse', '1 Test Way', 'Beirut', 'Lebanon', 'Test Manager', 1000, TRUE, TRUE),
    (2, 1001, 1001, 'Secondary Warehouse', '2 Test Way', 'Beirut', 'Lebanon', 'Test Manager', 500, FALSE, TRUE)
ON DUPLICATE KEY UPDATE business_id = VALUES(business_id), is_active = VALUES(is_active);

INSERT INTO stock_locations
    (id, owner_id, business_id, kind, source_table, source_id, name, is_default, is_active)
VALUES
    (8001, 1001, 1001, 'warehouse', 'native', 0, 'Default Warehouse', TRUE, TRUE),
    (8002, 1001, 1001, 'warehouse', 'erp_warehouses', 1, 'Main Warehouse', FALSE, TRUE),
    (8003, 1001, 1001, 'warehouse', 'erp_warehouses', 2, 'Secondary Warehouse', FALSE, TRUE)
ON DUPLICATE KEY UPDATE business_id = VALUES(business_id), is_active = VALUES(is_active);

INSERT INTO stock_levels
    (id, owner_id, business_id, location_id, product_id, on_hand, reorder_point, reorder_quantity)
VALUES
    (8101, 1001, 1001, 8001, 5201, 50.000, 10.000, 0.000),
    (8102, 1001, 1001, 8001, 5202, 200.000, 50.000, 0.000),
    (8103, 1001, 1001, 8001, 5203, 4.000, 10.000, 0.000),
    (8104, 1001, 1001, 8001, 5204, 75.000, 20.000, 0.000),
    (8111, 1001, 1001, 8002, 5201, 50.000, 10.000, 0.000),
    (8112, 1001, 1001, 8002, 5202, 200.000, 50.000, 0.000),
    (8113, 1001, 1001, 8002, 5203, 4.000, 10.000, 0.000),
    (8114, 1001, 1001, 8002, 5204, 75.000, 20.000, 0.000)
ON DUPLICATE KEY UPDATE on_hand = VALUES(on_hand), business_id = VALUES(business_id);

INSERT INTO stock_ledger
    (owner_id, business_id, location_id, product_id, entry_type, quantity_change, quantity_before, quantity_after, reference_type, reference_id, channel, note)
SELECT owner_id, business_id, location_id, product_id, 'opening', on_hand, 0, on_hand, 'fixture', id, 'erp', 'fixture opening balance'
FROM stock_levels
WHERE id BETWEEN 8101 AND 8104
ON DUPLICATE KEY UPDATE quantity_after = VALUES(quantity_after);

INSERT INTO stock_ledger
    (owner_id, business_id, location_id, product_id, entry_type, quantity_change, quantity_before, quantity_after, reference_type, reference_id, channel, note)
SELECT owner_id, business_id, location_id, product_id, 'opening', on_hand, 0, on_hand, 'fixture', id, 'erp', 'fixture opening balance'
FROM stock_levels
WHERE id BETWEEN 8111 AND 8114
ON DUPLICATE KEY UPDATE quantity_after = VALUES(quantity_after);
