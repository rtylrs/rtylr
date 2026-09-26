CREATE TABLE IF NOT EXISTS `authorization_permission_catalog` (
  `permission_key` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL,
  `app_id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `action` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `is_sensitive` tinyint(1) NOT NULL DEFAULT '0',
  `is_high_risk` tinyint(1) NOT NULL DEFAULT '0',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`permission_key`),
  KEY `idx_auth_permission_app` (`app_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `authorization_roles` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `business_id` bigint NOT NULL DEFAULT '0' COMMENT '0 means a global system role',
  `app_id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'dash',
  `role_key` varchar(96) COLLATE utf8mb4_unicode_ci NOT NULL,
  `display_name` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL,
  `description` text COLLATE utf8mb4_unicode_ci,
  `scope_rules` json DEFAULT NULL,
  `level` int NOT NULL DEFAULT '0',
  `parent_role_id` bigint DEFAULT NULL,
  `assignment_mode` enum('many','one','primary_plus_addons') COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'many',
  `is_system` tinyint(1) NOT NULL DEFAULT '0',
  `is_archived` tinyint(1) NOT NULL DEFAULT '0',
  `lifecycle` enum('draft','active','archived','replaced') COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'active',
  `version` bigint NOT NULL DEFAULT '1',
  `created_by` bigint DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_auth_role_scope_key` (`business_id`,`app_id`,`role_key`),
  KEY `idx_auth_roles_business_app` (`business_id`,`app_id`),
  KEY `idx_auth_roles_parent` (`parent_role_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `authorization_role_permissions` (
  `role_id` bigint NOT NULL,
  `permission_key` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`role_id`,`permission_key`),
  KEY `idx_auth_role_permissions_key` (`permission_key`),
  CONSTRAINT `fk_auth_role_permissions_permission` FOREIGN KEY (`permission_key`) REFERENCES `authorization_permission_catalog` (`permission_key`),
  CONSTRAINT `fk_auth_role_permissions_role` FOREIGN KEY (`role_id`) REFERENCES `authorization_roles` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `authorization_business_members` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `business_id` bigint NOT NULL,
  `user_id` bigint NOT NULL,
  `dash_role_id` bigint DEFAULT NULL,
  `dash_role_key` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'member',
  `status` enum('pending','active','suspended','revoked') COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'active',
  `authorization_version` bigint NOT NULL DEFAULT '1',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_auth_member_business_user` (`business_id`,`user_id`),
  KEY `idx_auth_member_user` (`user_id`),
  KEY `idx_auth_member_status` (`business_id`,`status`),
  CONSTRAINT `fk_auth_member_business` FOREIGN KEY (`business_id`) REFERENCES `businesses` (`id`),
  CONSTRAINT `fk_auth_member_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `authorization_license_assignments` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `business_id` bigint NOT NULL,
  `user_id` bigint DEFAULT NULL,
  `group_id` bigint DEFAULT NULL,
  `app_id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `source` enum('explicit','dash_default','group','system') COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'explicit',
  `status` enum('pending','active','revoked','archived') COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'active',
  `granted_by` bigint DEFAULT NULL,
  `granted_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `revoked_by` bigint DEFAULT NULL,
  `revoked_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_auth_active_direct_license` (`business_id`,`user_id`,`app_id`,`status`),
  KEY `idx_auth_license_user` (`business_id`,`user_id`,`status`),
  KEY `idx_auth_license_group` (`business_id`,`group_id`,`status`),
  KEY `idx_auth_license_app` (`business_id`,`app_id`,`status`),
  KEY `fk_auth_license_user` (`user_id`),
  KEY `fk_auth_license_group` (`group_id`),
  CONSTRAINT `fk_auth_license_business` FOREIGN KEY (`business_id`) REFERENCES `businesses` (`id`),
  CONSTRAINT `fk_auth_license_group` FOREIGN KEY (`group_id`) REFERENCES `user_groups` (`id`),
  CONSTRAINT `fk_auth_license_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`),
  CONSTRAINT `chk_auth_license_assignment_one_target` CHECK ((((`user_id` is not null) and (`group_id` is null)) or ((`user_id` is null) and (`group_id` is not null))))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `authorization_role_assignments` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `business_id` bigint NOT NULL,
  `user_id` bigint DEFAULT NULL,
  `group_id` bigint DEFAULT NULL,
  `app_id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `role_id` bigint NOT NULL,
  `source` enum('explicit','dash_default','group','system') COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'explicit',
  `status` enum('pending','active','revoked','archived') COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'active',
  `is_primary` tinyint(1) NOT NULL DEFAULT '0',
  `granted_by` bigint DEFAULT NULL,
  `granted_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `revoked_by` bigint DEFAULT NULL,
  `revoked_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_auth_active_direct_role` (`business_id`,`user_id`,`app_id`,`role_id`,`status`),
  KEY `idx_auth_role_assignment_user` (`business_id`,`user_id`,`app_id`,`status`),
  KEY `idx_auth_role_assignment_group` (`business_id`,`group_id`,`app_id`,`status`),
  KEY `idx_auth_role_assignment_role` (`role_id`,`status`),
  KEY `fk_auth_role_assignment_user` (`user_id`),
  KEY `fk_auth_role_assignment_group` (`group_id`),
  CONSTRAINT `fk_auth_role_assignment_business` FOREIGN KEY (`business_id`) REFERENCES `businesses` (`id`),
  CONSTRAINT `fk_auth_role_assignment_group` FOREIGN KEY (`group_id`) REFERENCES `user_groups` (`id`),
  CONSTRAINT `fk_auth_role_assignment_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`),
  CONSTRAINT `fk_auth_role_assignments_role` FOREIGN KEY (`role_id`) REFERENCES `authorization_roles` (`id`),
  CONSTRAINT `chk_auth_role_assignment_one_target` CHECK ((((`user_id` is not null) and (`group_id` is null)) or ((`user_id` is null) and (`group_id` is not null))))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `authorization_role_assignment_scopes` (
  `assignment_id` bigint NOT NULL,
  `scope_type` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `scope_id` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`assignment_id`,`scope_type`,`scope_id`),
  KEY `idx_auth_assignment_scope` (`scope_type`,`scope_id`),
  CONSTRAINT `fk_auth_assignment_scopes_assignment` FOREIGN KEY (`assignment_id`) REFERENCES `authorization_role_assignments` (`id`),
  CONSTRAINT `chk_auth_assignment_scope_type` CHECK (((`scope_type` in (_utf8mb4'pos_location',_utf8mb4'flow_project')) and regexp_like(`scope_id`,_utf8mb4'^[1-9][0-9]*$')))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `authorization_dash_app_defaults` (
  `dash_role_id` bigint NOT NULL,
  `app_id` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `default_role_id` bigint DEFAULT NULL,
  `default_enabled` tinyint(1) NOT NULL DEFAULT '0',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`dash_role_id`,`app_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `authorization_audit_events` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `business_id` bigint NOT NULL,
  `actor_user_id` bigint DEFAULT NULL,
  `target_user_id` bigint DEFAULT NULL,
  `target_group_id` bigint DEFAULT NULL,
  `action` varchar(96) COLLATE utf8mb4_unicode_ci NOT NULL,
  `app_id` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `role_id` bigint DEFAULT NULL,
  `before_json` json DEFAULT NULL,
  `after_json` json DEFAULT NULL,
  `reason` text COLLATE utf8mb4_unicode_ci,
  `approval_ref` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `request_id` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `correlation_id` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `source_client` enum('dash','pos','mobile','app','api') COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'api',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_auth_audit_business_time` (`business_id`,`created_at`),
  KEY `idx_auth_audit_target` (`business_id`,`target_user_id`,`created_at`),
  KEY `idx_auth_audit_action` (`business_id`,`action`,`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO authorization_permission_catalog
    (permission_key, app_id, action, is_sensitive, is_high_risk)
VALUES
    ('hub:dashboard:view', 'dash', 'view', FALSE, FALSE),
    ('hub:team:view', 'dash', 'view', FALSE, FALSE),
    ('hub:team:invite', 'dash', 'invite', FALSE, FALSE),
    ('hub:team:edit', 'dash', 'edit', FALSE, FALSE),
    ('hub:team:deactivate', 'dash', 'deactivate', FALSE, TRUE),
    ('hub:roles:view', 'dash', 'view', FALSE, FALSE),
    ('hub:roles:create', 'dash', 'create', FALSE, TRUE),
    ('hub:roles:edit', 'dash', 'edit', FALSE, TRUE),
    ('hub:roles:archive', 'dash', 'archive', FALSE, TRUE),
    ('hub:licenses:manage', 'dash', 'manage', FALSE, TRUE),
    ('hub:groups:manage', 'dash', 'manage', FALSE, FALSE),
    ('hub:hierarchy:manage', 'dash', 'manage', FALSE, TRUE),
    ('hub:audit:view', 'dash', 'view', FALSE, FALSE),
    ('hub:security:manage', 'dash', 'manage', TRUE, TRUE),
    ('hub:api_keys:manage', 'dash', 'manage', TRUE, TRUE),
    ('pos:orders:view', 'pos', 'view', FALSE, FALSE),
    ('pos:orders:create', 'pos', 'create', FALSE, FALSE),
    ('pos:orders:void', 'pos', 'void', FALSE, TRUE),
    ('pos:orders:void:approve', 'pos', 'approve', FALSE, TRUE),
    ('pos:orders:refund', 'pos', 'refund', FALSE, TRUE),
    ('pos:orders:refund:approve', 'pos', 'approve', FALSE, TRUE),
    ('pos:orders:void:override', 'pos', 'override', FALSE, TRUE),
    ('pos:orders:refund:override', 'pos', 'override', FALSE, TRUE),
    ('pos:orders:discount', 'pos', 'edit', FALSE, FALSE),
    ('pos:cash:view', 'pos', 'view', FALSE, FALSE),
    ('pos:cash:drop', 'pos', 'create', FALSE, TRUE),
    ('pos:cash:pickup', 'pos', 'create', FALSE, TRUE),
    ('pos:cash:approve', 'pos', 'approve', FALSE, TRUE),
    ('pos:employees:manage', 'pos', 'manage', FALSE, TRUE),
    ('pos:terminals:manage', 'pos', 'manage', FALSE, TRUE),
    ('pos:reports:view', 'pos', 'view', FALSE, FALSE),
    ('erp:products:view', 'erp', 'view', FALSE, FALSE),
    ('erp:products:create', 'erp', 'create', FALSE, FALSE),
    ('erp:products:edit', 'erp', 'edit', FALSE, FALSE),
    ('erp:products:delete', 'erp', 'delete', FALSE, TRUE),
    ('erp:inventory:view', 'erp', 'view', FALSE, FALSE),
    ('erp:inventory:adjust', 'erp', 'edit', FALSE, TRUE),
    ('erp:inventory:transfer', 'erp', 'edit', FALSE, TRUE),
    ('erp:inventory:approve', 'erp', 'approve', FALSE, TRUE),
    ('erp:warehouses:manage', 'erp', 'manage', FALSE, TRUE),
    ('erp:settings:view', 'erp', 'view', FALSE, FALSE),
    ('erp:settings:manage', 'erp', 'manage', FALSE, TRUE),
    ('erp:suppliers:view', 'erp', 'view', FALSE, FALSE),
    ('erp:suppliers:create', 'erp', 'create', FALSE, FALSE),
    ('erp:suppliers:edit', 'erp', 'edit', FALSE, TRUE),
    ('erp:purchasing:view', 'erp', 'view', FALSE, FALSE),
    ('erp:purchasing:create', 'erp', 'create', FALSE, FALSE),
    ('erp:purchasing:approve', 'erp', 'approve', FALSE, TRUE),
    ('finance:invoices:view', 'finance', 'view', FALSE, FALSE),
    ('finance:invoices:create', 'finance', 'create', FALSE, FALSE),
    ('finance:invoices:send', 'finance', 'send', FALSE, FALSE),
    ('finance:quotes:view', 'finance', 'view', FALSE, FALSE),
    ('finance:quotes:manage', 'finance', 'manage', FALSE, FALSE),
    ('finance:recurring_invoices:view', 'finance', 'view', FALSE, FALSE),
    ('finance:recurring_invoices:manage', 'finance', 'manage', FALSE, FALSE),
    ('finance:expenses:view', 'finance', 'view', FALSE, FALSE),
    ('finance:expenses:create', 'finance', 'create', FALSE, FALSE),
    ('finance:expenses:approve', 'finance', 'approve', FALSE, TRUE),
    ('finance:bank_feeds:view', 'finance', 'view', TRUE, FALSE),
    ('finance:bank_feeds:sensitive_view', 'finance', 'sensitive.view', TRUE, TRUE),
    ('finance:reconciliation:manage', 'finance', 'manage', FALSE, TRUE),
    ('finance:reports:view', 'finance', 'view', FALSE, FALSE),
    ('finance:reports:export', 'finance', 'export', FALSE, FALSE),
    ('finance:tax:configure', 'finance', 'configure', TRUE, TRUE),
    ('crm:customers:view', 'crm', 'view', FALSE, FALSE),
    ('crm:customers:create', 'crm', 'create', FALSE, FALSE),
    ('crm:customers:edit', 'crm', 'edit', FALSE, FALSE),
    ('crm:customers:delete', 'crm', 'delete', FALSE, TRUE),
    ('crm:customers:export', 'crm', 'export', TRUE, TRUE),
    ('crm:campaigns:view', 'crm', 'view', FALSE, FALSE),
    ('crm:campaigns:create', 'crm', 'create', FALSE, FALSE),
    ('crm:campaigns:send', 'crm', 'send', FALSE, FALSE),
    ('crm:loyalty:manage', 'crm', 'manage', FALSE, TRUE),
    ('crm:automations:manage', 'crm', 'manage', FALSE, TRUE),
    ('hr:employees:view', 'hr', 'view', FALSE, FALSE),
    ('hr:employees:manage', 'hr', 'manage', FALSE, TRUE),
    ('hr:employees:sensitive_view', 'hr', 'sensitive.view', TRUE, TRUE),
    ('hr:time_off:request', 'hr', 'create', FALSE, FALSE),
    ('hr:time_off:approve', 'hr', 'approve', FALSE, TRUE),
    ('hr:schedules:manage', 'hr', 'manage', FALSE, FALSE),
    ('hr:payroll:view', 'hr', 'view', TRUE, FALSE),
    ('hr:payroll:manage', 'hr', 'manage', TRUE, TRUE),
    ('hr:payroll:execute', 'hr', 'execute', TRUE, TRUE),
    ('hr:roles:manage', 'hr', 'manage', FALSE, TRUE),
    ('flow:projects:view', 'flow', 'view', FALSE, FALSE),
    ('flow:projects:create', 'flow', 'create', FALSE, FALSE),
    ('flow:projects:edit', 'flow', 'edit', FALSE, FALSE),
    ('flow:projects:assign', 'flow', 'assign', FALSE, TRUE),
    ('flow:tasks:view', 'flow', 'view', FALSE, FALSE),
    ('flow:tasks:create', 'flow', 'create', FALSE, FALSE),
    ('flow:tasks:edit', 'flow', 'edit', FALSE, FALSE),
    ('flow:tasks:assign', 'flow', 'assign', FALSE, TRUE),
    ('flow:team:manage', 'flow', 'manage', FALSE, TRUE),
    ('flow:managers:manage', 'flow', 'manage', FALSE, TRUE),
    ('flow:settings:manage', 'flow', 'manage', FALSE, TRUE),
    ('insights:dashboard:view', 'insights', 'view', FALSE, FALSE),
    ('insights:financial:view', 'insights', 'view', TRUE, FALSE),
    ('insights:customers:view', 'insights', 'view', TRUE, FALSE),
    ('insights:reports:create', 'insights', 'create', FALSE, FALSE),
    ('insights:reports:share', 'insights', 'share', FALSE, FALSE),
    ('insights:reports:export', 'insights', 'export', FALSE, FALSE),
    ('menu:menus:view', 'menu', 'view', FALSE, FALSE),
    ('menu:menus:edit', 'menu', 'edit', FALSE, FALSE),
    ('menu:branding:edit', 'menu', 'edit', FALSE, FALSE),
    ('menu:orders:view', 'menu', 'view', FALSE, FALSE),
    ('menu:analytics:view', 'menu', 'view', FALSE, FALSE),
    ('recruit:jobs:view', 'recruit', 'view', FALSE, FALSE),
    ('recruit:jobs:manage', 'recruit', 'manage', FALSE, TRUE),
    ('recruit:jobs:publish', 'recruit', 'publish', FALSE, TRUE),
    ('recruit:candidates:view', 'recruit', 'view', FALSE, FALSE),
    ('recruit:candidates:edit', 'recruit', 'edit', FALSE, FALSE),
    ('recruit:candidates:sensitive_view', 'recruit', 'sensitive.view', TRUE, TRUE),
    ('recruit:candidates:export', 'recruit', 'export', TRUE, TRUE)
ON DUPLICATE KEY UPDATE
    app_id = VALUES(app_id),
    action = VALUES(action),
    is_sensitive = VALUES(is_sensitive),
    is_high_risk = VALUES(is_high_risk);

INSERT INTO authorization_roles
    (business_id, app_id, role_key, display_name, description, level, assignment_mode, is_system)
VALUES
    (0, 'dash', 'dash.owner', 'Owner', 'Full tenant control.', 100, 'one', TRUE),
    (0, 'dash', 'dash.admin', 'Admin', 'Administrative tenant control.', 90, 'one', TRUE),
    (0, 'dash', 'dash.manager', 'Manager', 'Team and operational management.', 60, 'one', TRUE),
    (0, 'dash', 'dash.member', 'Member', 'Standard team membership.', 40, 'one', TRUE),
    (0, 'dash', 'dash.viewer', 'Viewer', 'Read-only tenant membership.', 10, 'one', TRUE),
    (0, 'pos', 'pos.viewer', 'POS Viewer', 'Read POS records.', 10, 'one', TRUE),
    (0, 'pos', 'pos.cashier', 'Cashier', 'Sell and manage assigned POS sessions.', 20, 'primary_plus_addons', TRUE),
    (0, 'pos', 'pos.senior_cashier', 'Senior Cashier', 'Cashier permissions plus approved exceptions.', 40, 'primary_plus_addons', TRUE),
    (0, 'pos', 'pos.location_manager', 'Location Manager', 'Manage POS operations for a location.', 70, 'one', TRUE),
    (0, 'pos', 'pos.admin', 'POS Admin', 'Full POS administration.', 90, 'one', TRUE),
    (0, 'erp', 'erp.viewer', 'ERP Viewer', 'Read ERP records.', 10, 'one', TRUE),
    (0, 'erp', 'erp.manager', 'ERP Manager', 'Manage ERP operations.', 60, 'one', TRUE),
    (0, 'erp', 'erp.admin', 'ERP Admin', 'Full ERP administration.', 90, 'one', TRUE),
    (0, 'hr', 'hr.viewer', 'HR Viewer', 'Read HR records.', 10, 'one', TRUE),
    (0, 'hr', 'hr.employee', 'Employee', 'View self-service HR records.', 10, 'one', TRUE),
    (0, 'hr', 'hr.manager', 'HR Manager', 'Manage HR records.', 60, 'one', TRUE),
    (0, 'hr', 'hr.admin', 'HR Admin', 'Full HR administration.', 90, 'one', TRUE),
    (0, 'crm', 'crm.viewer', 'CRM Viewer', 'Read CRM records.', 10, 'one', TRUE),
    (0, 'crm', 'crm.manager', 'CRM Manager', 'Manage CRM operations.', 60, 'one', TRUE),
    (0, 'crm', 'crm.admin', 'CRM Admin', 'Full CRM administration.', 90, 'one', TRUE),
    (0, 'finance', 'finance.viewer', 'Finance Viewer', 'Read finance records.', 10, 'one', TRUE),
    (0, 'finance', 'finance.manager', 'Finance Manager', 'Manage finance operations.', 60, 'one', TRUE),
    (0, 'finance', 'finance.admin', 'Finance Admin', 'Full finance administration.', 90, 'one', TRUE),
    (0, 'flow', 'flow.viewer', 'Flow Viewer', 'Read Flow workspaces.', 10, 'one', TRUE),
    (0, 'flow', 'flow.contributor', 'Flow Contributor', 'Create and edit assigned work.', 30, 'primary_plus_addons', TRUE),
    (0, 'flow', 'flow.team_leader', 'Flow Team Leader', 'Lead assigned Flow teams.', 50, 'one', TRUE),
    (0, 'flow', 'flow.manager', 'Flow Manager', 'Manage Flow projects and teams.', 70, 'one', TRUE),
    (0, 'flow', 'flow.admin', 'Flow Admin', 'Full Flow administration.', 90, 'one', TRUE),
    (0, 'insights', 'insights.viewer', 'Insights Viewer', 'Read Insights dashboards.', 10, 'one', TRUE),
    (0, 'insights', 'insights.analyst', 'Insights Analyst', 'Create and share Insights reports.', 50, 'one', TRUE),
    (0, 'insights', 'insights.admin', 'Insights Admin', 'Full Insights administration.', 90, 'one', TRUE),
    (0, 'menu', 'menu.viewer', 'Menu Viewer', 'Read menus and orders.', 10, 'one', TRUE),
    (0, 'menu', 'menu.editor', 'Menu Editor', 'Edit menus and branding.', 40, 'one', TRUE),
    (0, 'menu', 'menu.manager', 'Menu Manager', 'Manage menus and analytics.', 70, 'one', TRUE),
    (0, 'menu', 'menu.admin', 'Menu Admin', 'Full Menu administration.', 90, 'one', TRUE),
    (0, 'recruit', 'recruit.viewer', 'Recruit Viewer', 'Read recruiting records.', 10, 'one', TRUE),
    (0, 'recruit', 'recruit.recruiter', 'Recruiter', 'Manage recruiting workflows.', 60, 'one', TRUE),
    (0, 'recruit', 'recruit.manager', 'Recruit Manager', 'Manage recruiting operations.', 75, 'one', TRUE),
    (0, 'recruit', 'recruit.admin', 'Recruit Admin', 'Full Recruit administration.', 90, 'one', TRUE)
ON DUPLICATE KEY UPDATE
    display_name = VALUES(display_name),
    description = VALUES(description),
    level = VALUES(level),
    assignment_mode = VALUES(assignment_mode),
    is_system = TRUE,
    is_archived = FALSE,
    lifecycle = 'active';

INSERT INTO authorization_role_permissions (role_id, permission_key)
SELECT r.id, p.permission_key
FROM authorization_roles r
JOIN authorization_permission_catalog p
WHERE r.business_id = 0
  AND r.app_id = 'dash'
  AND r.role_key IN ('dash.owner', 'dash.admin')
ON DUPLICATE KEY UPDATE permission_key = VALUES(permission_key);

INSERT INTO authorization_role_permissions (role_id, permission_key)
SELECT r.id, p.permission_key
FROM authorization_roles r
JOIN authorization_permission_catalog p ON p.app_id = r.app_id
WHERE r.business_id = 0
  AND r.role_key LIKE CONCAT(r.app_id, '.admin')
ON DUPLICATE KEY UPDATE permission_key = VALUES(permission_key);

INSERT INTO authorization_role_permissions (role_id, permission_key)
SELECT r.id, p.permission_key
FROM authorization_roles r
JOIN authorization_permission_catalog p ON p.app_id = r.app_id
WHERE r.business_id = 0
  AND r.role_key IN (
      'pos.location_manager', 'erp.manager', 'hr.manager', 'crm.manager',
      'finance.manager', 'flow.manager', 'flow.team_leader',
      'insights.analyst', 'menu.manager', 'recruit.manager'
  )
  AND p.action IN ('view', 'create', 'edit', 'send', 'share', 'export')
ON DUPLICATE KEY UPDATE permission_key = VALUES(permission_key);

-- The generic manager grant above intentionally excludes high-risk `manage`
-- actions. Quote and recurring-invoice management are the two finance manager
-- capabilities that need that action at the manager tier.
INSERT INTO authorization_role_permissions (role_id, permission_key)
SELECT r.id, p.permission_key
FROM authorization_roles r
JOIN authorization_permission_catalog p ON p.app_id = r.app_id
WHERE r.business_id = 0
  AND r.role_key = 'finance.manager'
  AND p.permission_key IN ('finance:quotes:manage', 'finance:recurring_invoices:manage')
ON DUPLICATE KEY UPDATE permission_key = VALUES(permission_key);

INSERT INTO authorization_role_permissions (role_id, permission_key)
SELECT r.id, p.permission_key
FROM authorization_roles r
JOIN authorization_permission_catalog p ON p.app_id = r.app_id
WHERE r.business_id = 0
  AND r.role_key IN ('pos.cashier', 'flow.contributor', 'recruit.recruiter')
  AND p.action IN ('view', 'create', 'edit')
ON DUPLICATE KEY UPDATE permission_key = VALUES(permission_key);

INSERT INTO authorization_role_permissions (role_id, permission_key)
SELECT r.id, p.permission_key
FROM authorization_roles r
JOIN authorization_permission_catalog p ON p.app_id = r.app_id
WHERE r.business_id = 0
  AND r.role_key IN (
      'pos.senior_cashier', 'pos.viewer', 'erp.viewer', 'hr.employee',
      'hr.viewer', 'crm.viewer', 'finance.viewer', 'flow.viewer', 'insights.viewer',
      'menu.viewer', 'recruit.viewer'
  )
  AND p.action = 'view'
  AND p.is_sensitive = FALSE
ON DUPLICATE KEY UPDATE permission_key = VALUES(permission_key);
