-- Tenant + business fixtures.
-- One owner with one business gives every owner-scoped query something to
-- return without dragging in tenant data from a real run.

INSERT INTO businesses (id, name, owner_id, industry, country, timezone, currency, is_active)
VALUES
    (1, 'Test Business', 1001, 'restaurant', 'LB', 'Asia/Beirut', 'USD', TRUE),
    (1001, 'Test Business Owner Scope', 1001, 'restaurant', 'LB', 'UTC', 'USD', TRUE),
    (5100, 'Benefit API Business', 5201, 'services', 'LB', 'UTC', 'USD', TRUE)
ON DUPLICATE KEY UPDATE name = VALUES(name);

INSERT INTO authorization_business_members (business_id, user_id, member_type, status)
VALUES
    (1001, 1002, 'admin', 'active'),
    (1001, 1003, 'member', 'active'),
    (5100, 5202, 'member', 'active')
ON DUPLICATE KEY UPDATE member_type = VALUES(member_type), status = VALUES(status);

INSERT INTO subscriptions (business_id, plan_id, status)
SELECT business.id, plan.id, 'active'
FROM businesses business
JOIN plans plan ON plan.slug = 'free'
WHERE business.id IN (1, 1001)
ON DUPLICATE KEY UPDATE plan_id = VALUES(plan_id), status = VALUES(status);
