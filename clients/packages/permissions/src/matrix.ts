import type { PlatformRole } from "./roles";
import type { Resource } from "./capabilities";

// ── Permission matrix ──────────────────────────────────────────────
// Maps each role to the resources it can access.
// A role inherits all permissions from roles below it in the hierarchy.

type PermissionMatrix = Record<PlatformRole, Resource[]>;

/** Direct permissions per role (NOT inherited - use `getPermissions` for resolved set) */
const DIRECT_PERMISSIONS: PermissionMatrix = {
  viewer: ["hub:dashboard", "insights:dashboard"],

  staff: [
    "app:pos",
    "app:flow",
    "app:hr",
    "pos:orders",
    "pos:orders:create",
    "pos:tables",
    "pos:sessions",
    "pos:sessions:open",
    "pos:sessions:close",
    "pos:kds",
    "pos:receipts",
    "pos:drawer:open",
    "flow:tasks",
    "flow:projects",
    // Self-service only. Their own shifts and their own time-off requests are
    // things a person needs about themselves; hr:staff is the staff directory,
    // which is a different kind of access and now starts at operator (#795).
    "hr:shifts",
    "hr:time_off",
  ],

  // store_manager has NO extra GLOBAL permissions: its elevated POS powers
  // (manage staff PINs, voids, refunds, reports) are location-scoped and
  // enforced per-location via pos_member_locations + the capability flags
  // returned at PIN verify. Keeping this empty also prevents those powers from
  // leaking upward to menu_manager/billing through hierarchy inheritance.
  store_manager: [],

  menu_manager: ["app:menu", "hub:design"],

  billing: [
    "hub:billing",
    "finance:invoices",
    "finance:quotes:view",
    "finance:quotes:manage",
    "finance:recurring_invoices:view",
    "finance:recurring_invoices:manage",
    "finance:expenses",
    "finance:reports",
  ],

  operator: [
    "app:erp",
    "app:crm",
    "app:insights",
    // These five were never declared for operator. They arrived because
    // menu_manager and billing sat below operator in the old flat list, so a
    // model change would have silently revoked them. Declared explicitly to
    // keep operator's access exactly as it was (#795).
    //
    // Worth reviewing separately: hub:billing on a mid-tier operator is a
    // consequence of the old ordering, not a decision anyone recorded.
    "app:menu",
    "hub:design",
    "hub:billing",
    "finance:invoices",
    "finance:quotes:view",
    "finance:recurring_invoices:view",
    "finance:reports",
    // The staff directory, previously reached through staff's own grant.
    "hr:staff",
    "hub:team",
    "hub:activity",
    "pos:orders:hold",
    "pos:orders:discount",
    "pos:cash:drop",
    "pos:reports",
    "pos:employees",
    "erp:products",
    "erp:products:create",
    "erp:products:edit",
    "erp:inventory",
    "erp:suppliers",
    "erp:purchase_orders",
    "crm:customers",
    "crm:customers:manage",
    "crm:campaigns",
    "crm:segments",
    "crm:loyalty",
    "hr:shifts:manage",
    "hr:departments",
    "finance:invoices:create",
    "finance:expenses",
    "flow:projects:create",
    "flow:tasks:assign",
    "flow:views",
    "insights:reports",
  ],

  admin: [
    "hub:settings",
    "hub:addons",
    "hub:developer",
    "hub:integrations",
    "hub:permissions",
    "hub:team:invite",
    "hub:team:remove",
    "pos:orders:void",
    "pos:orders:refund",
    "pos:cash:pickup",
    "pos:employees:manage",
    "pos:terminals",
    "pos:terminals:manage",
    "erp:products:delete",
    "erp:inventory:adjust",
    "erp:inventory:transfer",
    "erp:settings:manage",
    "erp:suppliers:create",
    "erp:suppliers:edit",
    "erp:purchase_orders:approve",
    "erp:stock_takes",
    "hr:staff:manage",
    "hr:payroll",
    "hr:time_off:approve",
    "hr:reviews",
    "crm:campaigns:send",
    "crm:loyalty:configure",
    "crm:automations",
    "finance:invoices:send",
    "finance:quotes:manage",
    "finance:recurring_invoices:manage",
    "finance:expenses:approve",
    "finance:tax",
    "finance:accounts",
    "finance:reconcile",
    "flow:roadmap",
    "insights:reports:export",
  ],

  owner: [
    "hr:payroll:run",
    // Owner inherits everything - this list is for owner-only actions
  ],
};

// ── Build resolved permission sets (with inheritance) ───────────────

import { ROLE_PARENT } from "./roles";

const resolvedCache = new Map<PlatformRole, Set<Resource>>();

/**
 * All permissions for a role: its own, plus its base role's, following
 * ROLE_PARENT to the end of the chain.
 *
 * This walks a chain rather than accumulating over an array index, so a role
 * only ever inherits from the base it names. Before #795 the index form meant
 * every role inherited everything declared below it in one flat list, which is
 * how a finance role acquired the cash drawer.
 */
export function getPermissions(role: PlatformRole): Set<Resource> {
  if (resolvedCache.has(role)) return resolvedCache.get(role)!;

  const permissions = new Set<Resource>();
  const seen = new Set<PlatformRole>();

  for (
    let current: PlatformRole | null = role;
    current !== null;
    current = ROLE_PARENT[current]
  ) {
    // ROLE_PARENT is authored by hand, so guard the cycle rather than hang.
    if (seen.has(current)) {
      throw new Error(
        `ROLE_PARENT contains a cycle reachable from "${role}" at "${current}"`,
      );
    }
    seen.add(current);
    for (const perm of DIRECT_PERMISSIONS[current]) {
      permissions.add(perm);
    }
  }

  resolvedCache.set(role, permissions);
  return permissions;
}

/** Check if a role has a specific permission */
export function hasPermission(role: PlatformRole, resource: Resource): boolean {
  return getPermissions(role).has(resource);
}

/** Check if a role has ALL of the specified permissions */
export function hasAllPermissions(
  role: PlatformRole,
  resources: Resource[],
): boolean {
  const perms = getPermissions(role);
  return resources.every((r) => perms.has(r));
}

/** Check if a role has ANY of the specified permissions */
export function hasAnyPermission(
  role: PlatformRole,
  resources: Resource[],
): boolean {
  const perms = getPermissions(role);
  return resources.some((r) => perms.has(r));
}
