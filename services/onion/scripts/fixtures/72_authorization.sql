-- A non-admin member with app licenses and permission profiles for resolver tests.
-- Flow deliberately has no project scope, so its resource checks deny access.

INSERT INTO authorization_business_members
    (id, business_id, user_id, member_type, status, authorization_version)
VALUES
    (9000000001, 1, 1002, 'member', 'active', 1)
ON DUPLICATE KEY UPDATE
    member_type = VALUES(member_type),
    status = VALUES(status),
    authorization_version = VALUES(authorization_version);

INSERT INTO authorization_license_assignments
    (id, business_id, user_id, app_id, source, status, granted_by)
VALUES
    (9000000010, 1, 1002, 'crm',  'explicit', 'active', 1001),
    (9000000011, 1, 1002, 'hr',   'explicit', 'active', 1001),
    (9000000012, 1, 1002, 'flow', 'explicit', 'active', 1001)
ON DUPLICATE KEY UPDATE status = VALUES(status);

SET @crm_viewer_profile := (
    SELECT id FROM authorization_roles
    WHERE business_id = 0 AND role_key = 'crm.viewer' LIMIT 1
);
SET @hr_viewer_profile := (
    SELECT id FROM authorization_roles
    WHERE business_id = 0 AND role_key = 'hr.viewer' LIMIT 1
);
SET @flow_viewer_profile := (
    SELECT id FROM authorization_roles
    WHERE business_id = 0 AND role_key = 'flow.viewer' LIMIT 1
);

INSERT INTO authorization_role_assignments
    (id, business_id, user_id, app_id, role_id, source, status, is_primary, granted_by)
VALUES
    (9000000020, 1, 1002, 'crm',  @crm_viewer_profile,  'explicit', 'active', TRUE, 1001),
    (9000000021, 1, 1002, 'hr',   @hr_viewer_profile,   'explicit', 'active', TRUE, 1001),
    (9000000022, 1, 1002, 'flow', @flow_viewer_profile, 'explicit', 'active', TRUE, 1001)
ON DUPLICATE KEY UPDATE status = VALUES(status), role_id = VALUES(role_id);
