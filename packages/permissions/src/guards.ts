import type { PlatformRole } from "./roles";
import type { POSRole } from "./roles";
import type { Resource } from "./capabilities";
import { hasPermission } from "./matrix";
import type {
  EffectiveAccess,
  PermissionKey,
  RoleScope,
  TenantAppId,
} from "@voxire/types";

// ── App access guard ────────────────────────────────────────────────

export type AppId =
  | "dash"
  | "pos"
  | "erp"
  | "hr"
  | "crm"
  | "finance"
  | "flow"
  | "insights"
  | "menu"
  | "recruit";

const APP_RESOURCE_MAP: Record<AppId, Resource> = {
  dash: "hub:dashboard",
  pos: "app:pos",
  erp: "app:erp",
  hr: "app:hr",
  crm: "app:crm",
  finance: "finance:invoices",
  flow: "app:flow",
  insights: "app:insights",
  menu: "app:menu",
  recruit: "app:recruit",
};

/** Check if a platform role can access a specific app */
export function canAccessApp(role: PlatformRole, appId: AppId): boolean {
  // Owner and admin can access everything
  if (role === "owner" || role === "admin") return true;
  const resource = APP_RESOURCE_MAP[appId];
  if (!resource) return false;
  return hasPermission(role, resource);
}

/** Get list of apps a role can access */
export function getAccessibleApps(role: PlatformRole): AppId[] {
  return (Object.keys(APP_RESOURCE_MAP) as AppId[]).filter((appId) =>
    canAccessApp(role, appId),
  );
}

// ── POS action guards ───────────────────────────────────────────────

/** Check if a POS user can perform a specific action */
export function canPOSAction(
  posRole: POSRole,
  action:
    | "void_order"
    | "refund_order"
    | "apply_discount"
    | "open_drawer"
    | "manage_employees"
    | "manage_terminals"
    | "view_reports"
    | "cash_drop"
    | "cash_pickup",
): boolean {
  switch (action) {
    case "void_order":
    case "refund_order":
    case "manage_employees":
    case "manage_terminals":
    case "view_reports":
    case "cash_pickup":
      return posRole === "manager";
    case "apply_discount":
    case "cash_drop":
      return posRole === "manager" || posRole === "cashier";
    case "open_drawer":
      return (
        posRole === "manager" || posRole === "cashier" || posRole === "server"
      );
    default:
      return false;
  }
}

// ── Generic permission check ────────────────────────────────────────

export interface PermissionContext {
  platformRole: PlatformRole;
  posRole?: POSRole;
  assignedApps?: string[];
  assignedLocationIds?: number[];
  effectiveAccess?: EffectiveAccess;
}

/** Universal permission check - combines platform role, POS role, and app assignments */
export function can(ctx: PermissionContext, resource: Resource): boolean {
  if (ctx.effectiveAccess) {
    return canEffectiveAccess(ctx.effectiveAccess, resource);
  }

  // Owner bypasses everything
  if (ctx.platformRole === "owner") return true;

  // Check platform permission
  if (hasPermission(ctx.platformRole, resource)) return true;

  // If they have explicit app assignments, check those
  if (ctx.assignedApps && ctx.assignedApps.length > 0) {
    const appId = resource.split(":")[0] as string;
    if (appId === "app") {
      const app = resource.split(":")[1];
      return ctx.assignedApps.includes(app);
    }
  }

  return false;
}

const APP_IDS: readonly TenantAppId[] = [
  "pos",
  "erp",
  "hr",
  "crm",
  "finance",
  "flow",
  "insights",
  "menu",
  "recruit",
];

function appForPermission(permission: string): TenantAppId | null {
  const parts = permission.split(":");
  const prefix = parts[0] === "app" ? parts[1] : parts[0];
  return APP_IDS.includes(prefix as TenantAppId)
    ? (prefix as TenantAppId)
    : null;
}

function appAccessFor(access: EffectiveAccess, appId: TenantAppId) {
  return access.apps.find((app) => app.appId === appId);
}

/**
 * Evaluates the server-resolved access document. Missing app or permission
 * data is denied. App licenses are checked independently from role grants.
 */
export function canEffectiveAccess(
  access: EffectiveAccess | null | undefined,
  permission: PermissionKey,
  appId?: TenantAppId,
): boolean {
  if (!access) return false;
  if (permission.startsWith("hub:"))
    return access.dashPermissions.includes(permission);

  const resolvedAppId = appId ?? appForPermission(permission);
  if (!resolvedAppId) return false;
  const appAccess = appAccessFor(access, resolvedAppId);
  if (!appAccess?.licensed || !appAccess.hasAccess) return false;
  if (permission === `app:${resolvedAppId}`) return true;
  return appAccess.permissions.includes(permission);
}

const LEGACY_PERMISSION_ALIASES: Record<string, PermissionKey> = {
  "hub:team:remove": "hub:team:deactivate",
  "hub:permissions": "hub:roles:view",
  "hub:settings": "hub:team:edit",
  "pos:orders": "pos:orders:view",
  "pos:reports": "pos:reports:view",
  "pos:employees": "pos:employees:manage",
  "pos:terminals": "pos:terminals:manage",
  "erp:products": "erp:products:view",
  "erp:suppliers": "erp:suppliers:view",
  "erp:purchase_orders": "erp:purchasing:view",
  "erp:purchase_orders:approve": "erp:purchasing:approve",
  "erp:stock_takes": "erp:inventory:adjust",
  "hr:staff": "hr:employees:view",
  "hr:staff:manage": "hr:employees:manage",
  "hr:time_off": "hr:time_off:request",
  "hr:departments": "hr:employees:view",
  "hr:reviews": "hr:employees:manage",
  "hr:payroll": "hr:payroll:view",
  "hr:payroll:run": "hr:payroll:execute",
  "hr:shifts": "hr:schedules:manage",
  "hr:shifts:manage": "hr:schedules:manage",
  "crm:customers": "crm:customers:view",
  "crm:customers:manage": "crm:customers:edit",
  "crm:campaigns": "crm:campaigns:view",
  "crm:loyalty": "crm:loyalty:manage",
  "crm:loyalty:configure": "crm:loyalty:manage",
  "crm:segments": "crm:customers:view",
  "crm:automations": "crm:automations:manage",
  "finance:invoices": "finance:invoices:view",
  "finance:quotes": "finance:quotes:view",
  "finance:recurring_invoices": "finance:recurring_invoices:view",
  "finance:expenses": "finance:expenses:view",
  "finance:reports": "finance:reports:view",
  "finance:tax": "finance:tax:configure",
  "finance:accounts": "finance:bank_feeds:view",
  "finance:reconcile": "finance:reconciliation:manage",
  "flow:projects": "flow:projects:view",
  "flow:tasks": "flow:tasks:view",
  "flow:views": "flow:projects:view",
  "flow:roadmap": "flow:projects:view",
  "insights:dashboard": "insights:dashboard:view",
  "insights:reports": "insights:reports:create",
  "insights:reports:export": "insights:reports:export",
};

export function canonicalizePermission(permission: string): PermissionKey {
  return LEGACY_PERMISSION_ALIASES[permission] ?? permission;
}

/** Checks an app action and its required row scope. */
export function canEffectiveAccessInScope(
  access: EffectiveAccess | null | undefined,
  permission: PermissionKey,
  scope?: RoleScope,
  appId?: TenantAppId,
): boolean {
  if (!canEffectiveAccess(access, permission, appId)) return false;
  if (!scope) return true;
  const resolvedAppId = appId ?? appForPermission(permission);
  if (!resolvedAppId || !access) return false;
  const appAccess = appAccessFor(access, resolvedAppId);
  if (!appAccess) return false;
  return appAccess.scopes.some(
    (candidate) =>
      candidate.scopeType === "business" ||
      (candidate.scopeType === scope.scopeType &&
        candidate.scopeId === scope.scopeId),
  );
}

/** Check multiple permissions at once */
export function canAll(ctx: PermissionContext, resources: Resource[]): boolean {
  return resources.every((r) => can(ctx, r));
}

export function canAny(ctx: PermissionContext, resources: Resource[]): boolean {
  return resources.some((r) => can(ctx, r));
}
