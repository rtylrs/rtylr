/** Applications that participate in tenant employee authorization. */
export type TenantAppId =
  | "pos"
  | "erp"
  | "hr"
  | "crm"
  | "finance"
  | "flow"
  | "insights"
  | "menu"
  | "recruit";

export type AuthorizationAppId = "dash" | TenantAppId;

export type DashRole = "owner" | "admin" | "manager" | "member" | "viewer";

export type AssignmentSource = "explicit" | "dash_default" | "group" | "system";

export type AssignmentStatus = "pending" | "active" | "revoked" | "archived";

export type AssignmentMode = "many" | "one" | "primary_plus_addons";

export type RoleScopeType =
  | "business"
  | "location"
  | "department"
  | "team"
  | "project"
  | "terminal"
  | "portfolio";

export interface RoleScope {
  scopeType: RoleScopeType;
  scopeId: string;
}

export type PermissionAction =
  | "view"
  | "create"
  | "edit"
  | "delete"
  | "archive"
  | "approve"
  | "publish"
  | "send"
  | "export"
  | "assign"
  | "manage"
  | "configure"
  | "execute"
  | "override"
  | "sensitive.view";

export const KNOWN_PERMISSION_KEYS = [
  "hub:dashboard:view",
  "hub:team:view",
  "hub:team:invite",
  "hub:team:edit",
  "hub:team:deactivate",
  "hub:roles:view",
  "hub:roles:create",
  "hub:roles:edit",
  "hub:roles:archive",
  "hub:licenses:manage",
  "hub:groups:manage",
  "hub:hierarchy:manage",
  "hub:audit:view",
  "hub:security:manage",
  "hub:api_keys:manage",
  "pos:orders:view",
  "pos:orders:create",
  "pos:orders:void",
  "pos:orders:void:approve",
  "pos:orders:refund",
  "pos:orders:refund:approve",
  "pos:orders:void:override",
  "pos:orders:refund:override",
  "pos:orders:discount",
  "pos:cash:view",
  "pos:cash:drop",
  "pos:cash:pickup",
  "pos:cash:approve",
  "pos:employees:manage",
  "pos:terminals:manage",
  "pos:reports:view",
  "erp:products:view",
  "erp:products:create",
  "erp:products:edit",
  "erp:products:delete",
  "erp:inventory:view",
  "erp:inventory:adjust",
  "erp:inventory:transfer",
  "erp:inventory:approve",
  "erp:warehouses:manage",
  "erp:settings:view",
  "erp:settings:manage",
  "erp:suppliers:view",
  "erp:suppliers:create",
  "erp:suppliers:edit",
  "erp:purchasing:view",
  "erp:purchasing:create",
  "erp:purchasing:approve",
  "finance:invoices:view",
  "finance:invoices:create",
  "finance:invoices:send",
  "finance:quotes:view",
  "finance:quotes:manage",
  "finance:recurring_invoices:view",
  "finance:recurring_invoices:manage",
  "finance:expenses:view",
  "finance:expenses:create",
  "finance:expenses:approve",
  "finance:bank_feeds:view",
  "finance:bank_feeds:sensitive_view",
  "finance:reconciliation:manage",
  "finance:reports:view",
  "finance:reports:export",
  "finance:tax:configure",
  "crm:customers:view",
  "crm:customers:create",
  "crm:customers:edit",
  "crm:customers:delete",
  "crm:customers:export",
  "crm:campaigns:view",
  "crm:campaigns:create",
  "crm:campaigns:send",
  "crm:loyalty:manage",
  "crm:automations:manage",
  "hr:employees:view",
  "hr:employees:manage",
  "hr:employees:sensitive_view",
  "hr:time_off:request",
  "hr:time_off:approve",
  "hr:schedules:manage",
  "hr:payroll:view",
  "hr:payroll:manage",
  "hr:payroll:execute",
  "hr:roles:manage",
  "flow:projects:view",
  "flow:projects:create",
  "flow:projects:edit",
  "flow:projects:assign",
  "flow:tasks:view",
  "flow:tasks:create",
  "flow:tasks:edit",
  "flow:tasks:assign",
  "flow:team:manage",
  "flow:managers:manage",
  "flow:settings:manage",
  "insights:dashboard:view",
  "insights:financial:view",
  "insights:customers:view",
  "insights:reports:create",
  "insights:reports:share",
  "insights:reports:export",
  "menu:menus:view",
  "menu:menus:edit",
  "menu:branding:edit",
  "menu:orders:view",
  "menu:analytics:view",
  "recruit:jobs:view",
  "recruit:jobs:manage",
  "recruit:jobs:publish",
  "recruit:candidates:view",
  "recruit:candidates:edit",
  "recruit:candidates:sensitive_view",
  "recruit:candidates:export",
] as const;

/** Open extension type for business-specific permissions, while preserving catalog literals. */
export type PermissionKey =
  (typeof KNOWN_PERMISSION_KEYS)[number] | (string & {});

export interface RoleDefinition {
  id: string;
  businessId: string | null;
  appId: AuthorizationAppId;
  key: string;
  name: string;
  description: string;
  level: number;
  parentRoleId: string | null;
  assignmentMode: AssignmentMode;
  permissions: PermissionKey[];
  isSystem: boolean;
  isArchived: boolean;
  version: number;
}

export interface AppRoleAssignment {
  appId: TenantAppId;
  roleIds: string[];
  primaryRoleId: string | null;
  source: AssignmentSource;
  status: AssignmentStatus;
  scopes: RoleScope[];
  groupId: string | null;
}

export interface LicenseAssignment {
  appId: TenantAppId;
  status: AssignmentStatus;
  source: AssignmentSource;
  groupId: string | null;
  grantedAt: string;
  revokedAt: string | null;
}

export interface EffectiveAppAccess {
  appId: TenantAppId;
  licensed: boolean;
  hasAccess: boolean;
  roleIds: string[];
  primaryRoleId: string | null;
  displayRole: string | null;
  permissions: PermissionKey[];
  scopes: RoleScope[];
  sources: AssignmentSource[];
  permissionSources: Record<string, string[]>;
}

export interface EffectiveAccess {
  businessId: string;
  dashRole: DashRole | string | null;
  dashRoleId: string | null;
  dashPermissions: PermissionKey[];
  subscribedApps: TenantAppId[];
  licenses: LicenseAssignment[];
  apps: EffectiveAppAccess[];
  isOwner: boolean;
  isAdmin: boolean;
  authorizationVersion: number;
}

export type ProvisioningMode = "invitation" | "direct_credentials";

export type ProvisioningState =
  | "pending_invitation"
  | "pending_activation"
  | "active"
  | "suspended"
  | "revoked";

export interface ProvisioningStatus {
  mode: ProvisioningMode;
  state: ProvisioningState;
  emailVerified: boolean;
  passwordChangeRequired: boolean;
  mfaRequired: boolean;
  mfaEnrolled: boolean;
  mfaVerified: boolean;
  recoveryAvailable: boolean;
}

export interface AuthorizationAuditEvent {
  id: string;
  businessId: string;
  actorUserId: string;
  targetUserId: string | null;
  targetGroupId: string | null;
  action: string;
  appId: AuthorizationAppId | null;
  roleId: string | null;
  before: Record<string, unknown>;
  after: Record<string, unknown>;
  reason: string | null;
  sourceClient: "dash" | "pos" | "mobile" | "app" | "api" | null;
  createdAt: string;
}
