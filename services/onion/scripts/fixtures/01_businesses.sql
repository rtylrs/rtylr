-- Tenant + business fixtures.
-- One owner with one business gives every owner-scoped query something to
-- return without dragging in tenant data from a real run.

INSERT INTO businesses (id, name, owner_id, industry, country, timezone, currency, is_active)
VALUES
    (1, 'Test Business', 1001, 'restaurant', 'LB', 'Asia/Beirut', 'USD', TRUE)
ON DUPLICATE KEY UPDATE name = VALUES(name);
