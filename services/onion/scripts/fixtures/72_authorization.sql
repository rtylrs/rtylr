-- Normalized role-access v2 rows for a real non-owner member.
--
-- The ordinary fixtures only ever created legacy `user_app_permissions` rows,
-- so the normalized member shape was exercised by unit tests with fabricated
-- state and never against a seeded database. Two bugs hid in that gap: an
-- unscoped HR assignment was denied by the scope gate, and the legacy apps
-- column was the only entitlement source the resolver saw. This file seeds the
-- state the product can actually produce, so both are covered by the real
-- backend suite.
--
-- Ids follow 70_direct_authorization.sql: 9xxxxxxx for authorization rows, and
-- user 1002 (staff@test.local, business 1) as the non-owner member.
--
-- System role ids referenced below come from zz_role_access_v2.sql:
--   4 = dash.member, 15 = hr.employee, 18 = crm.viewer, 24 = flow.viewer.

-- Membership: active and normalized (authorization_version = 1). The version
-- matters, because 0 is the legacy migration escape hatch that bypasses both the
-- scope rule and the assignment requirement.
INSERT INTO authorization_business_members
    (id, business_id, user_id, dash_role_id, dash_role_key, status, authorization_version)
VALUES
    (9000000001, 1, 1002, 4, 'dash.member', 'active', 1)
ON DUPLICATE KEY UPDATE
    dash_role_id = VALUES(dash_role_id),
    dash_role_key = VALUES(dash_role_key),
    status = VALUES(status),
    authorization_version = VALUES(authorization_version);

-- Licenses are plan metadata: they are reported, never enforced. They are seeded
-- so the plan state looks realistic, but access is decided by the assignments
-- below, and a self-hosted install can legitimately have no license rows.
INSERT INTO authorization_license_assignments
    (id, business_id, user_id, app_id, source, status, granted_by)
VALUES
    (9000000010, 1, 1002, 'crm',  'explicit', 'active', 1001),
    (9000000011, 1, 1002, 'hr',   'explicit', 'active', 1001),
    (9000000012, 1, 1002, 'flow', 'explicit', 'active', 1001)
ON DUPLICATE KEY UPDATE status = VALUES(status);

-- Two app shapes side by side.
--
-- crm and hr are unscoped: neither app models a resource scope, and the CHECK
-- constraint permits no scope type that would apply to them. The HR row is the
-- exact state that used to lock every non-owner member out of the app.
--
-- flow is deliberately left unscoped too. Flow *is* a resource scoped app, so
-- this assignment must stay denied until it carries a flow_project scope. No
-- flow project is seeded for business 1, and ValidateAssignmentScope would
-- correctly reject a scope pointing at a row that does not exist.
INSERT INTO authorization_role_assignments
    (id, business_id, user_id, app_id, role_id, source, status, is_primary, granted_by)
VALUES
    (9000000020, 1, 1002, 'crm',  18, 'explicit', 'active', TRUE, 1001),
    (9000000021, 1, 1002, 'hr',   15, 'explicit', 'active', TRUE, 1001),
    (9000000022, 1, 1002, 'flow', 24, 'explicit', 'active', TRUE, 1001)
ON DUPLICATE KEY UPDATE status = VALUES(status), role_id = VALUES(role_id);

-- The member's own legacy row, which must NOT be what grants access: emptying
-- it is the regression check for the entitlement source. `pos` is named here on
-- purpose, and on no assignment above, so a stale read of this column would hand
-- the member an app that no role backs.
INSERT INTO user_app_permissions
    (business_id, owner_id, user_id, apps, role, is_active)
VALUES
    (1, 1001, 1002, '["dash", "crm", "hr", "flow", "pos"]', 'staff', TRUE)
ON DUPLICATE KEY UPDATE apps = VALUES(apps), is_active = VALUES(is_active);
