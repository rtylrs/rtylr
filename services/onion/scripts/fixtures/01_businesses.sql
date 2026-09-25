-- Tenant + business fixtures.
-- One owner with one business gives every owner-scoped query something to
-- return without dragging in tenant data from a real run.

INSERT INTO businesses (id, name, owner_id, industry, country, timezone, currency, is_active)
VALUES
    (1, 'Test Business', 1001, 'restaurant', 'LB', 'Asia/Beirut', 'USD', TRUE),
    (1001, 'Test Business Owner Scope', 1001, 'restaurant', 'LB', 'UTC', 'USD', TRUE),
    (5100, 'Benefit API Business', 5201, 'services', 'LB', 'UTC', 'USD', TRUE)
ON DUPLICATE KEY UPDATE name = VALUES(name);

INSERT INTO user_business_access (user_id, business_id, role)
VALUES
    (1001, 1001, 'owner'),
    (1002, 1001, 'admin'),
    (1003, 1001, 'viewer'),
    (5201, 5100, 'owner'),
    (5202, 5100, 'owner')
ON DUPLICATE KEY UPDATE role = VALUES(role);
