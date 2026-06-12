import type { PlatformRole } from './roles';
import type { Resource } from './capabilities';

// ── Permission matrix ──────────────────────────────────────────────
// Maps each role to the resources it can access.
// A role inherits all permissions from roles below it in the hierarchy.

type PermissionMatrix = Record<PlatformRole, Resource[]>;

/** Direct permissions per role (NOT inherited - use `getPermissions` for resolved set) */
const DIRECT_PERMISSIONS: PermissionMatrix = {
  viewer: ['hub:dashboard', 'insights:dashboard'],

  staff: [
    'app:pos',
    'app:flow',
    'app:hr',
    'pos:orders',
    'pos:orders:create',
    'pos:tables',
    'pos:sessions',
    'pos:sessions:open',
    'pos:sessions:close',
    'pos:kds',
    'pos:receipts',
    'pos:drawer:open',
    'flow:tasks',
    'flow:projects',
    'hr:staff',
    'hr:shifts',
    'hr:time_off',
  ],

  // store_manager has NO extra GLOBAL permissions: its elevated POS powers
  // (manage staff PINs, voids, refunds, reports) are location-scoped and
  // enforced per-location via pos_member_locations + the capability flags
  // returned at PIN verify. Keeping this empty also prevents those powers from
  // leaking upward to menu_manager/billing through hierarchy inheritance.
  store_manager: [],

  menu_manager: ['app:menu', 'hub:design'],

  billing: ['hub:billing', 'finance:invoices', 'finance:expenses', 'finance:reports'],

  operator: [
    'app:erp',
    'app:crm',
    'app:insights',
    'hub:team',
    'hub:activity',
    'pos:orders:hold',
    'pos:orders:discount',
    'pos:cash:drop',
    'pos:reports',
    'pos:employees',
    'erp:products',
    'erp:products:create',
    'erp:products:edit',
    'erp:inventory',
    'erp:suppliers',
    'erp:purchase_orders',
    'crm:customers',
    'crm:customers:manage',
    'crm:campaigns',
    'crm:segments',
    'crm:loyalty',
    'hr:shifts:manage',
    'hr:departments',
    'finance:invoices:create',
    'finance:expenses',
    'flow:projects:create',
    'flow:tasks:assign',
    'flow:views',
    'insights:reports',
  ],

  admin: [
    'hub:settings',
    'hub:addons',
    'hub:developer',
    'hub:integrations',
    'hub:permissions',
    'hub:team:invite',
    'hub:team:remove',
    'pos:orders:void',
    'pos:orders:refund',
    'pos:cash:pickup',
    'pos:employees:manage',
    'pos:terminals',
    'pos:terminals:manage',
    'erp:products:delete',
    'erp:inventory:adjust',
    'erp:inventory:transfer',
    'erp:stock_takes',
    'hr:staff:manage',
    'hr:payroll',
    'hr:time_off:approve',
    'hr:reviews',
    'crm:campaigns:send',
    'crm:loyalty:configure',
    'crm:automations',
    'finance:invoices:send',
    'finance:expenses:approve',
    'finance:tax',
    'finance:accounts',
    'finance:reconcile',
    'flow:roadmap',
    'insights:reports:export',
  ],

  owner: [
    'hr:payroll:run',
    // Owner inherits everything - this list is for owner-only actions
  ],
};

// ── Build resolved permission sets (with inheritance) ───────────────

import { ROLE_HIERARCHY } from './roles';

const resolvedCache = new Map<PlatformRole, Set<Resource>>();

/** Get all permissions for a role, including inherited from lower roles */
export function getPermissions(role: PlatformRole): Set<Resource> {
  if (resolvedCache.has(role)) return resolvedCache.get(role)!;

  const roleIndex = ROLE_HIERARCHY.indexOf(role);
  const permissions = new Set<Resource>();

  // Accumulate from lowest role up to this role
  for (let i = 0; i <= roleIndex; i++) {
    const r = ROLE_HIERARCHY[i];
    for (const perm of DIRECT_PERMISSIONS[r]) {
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
export function hasAllPermissions(role: PlatformRole, resources: Resource[]): boolean {
  const perms = getPermissions(role);
  return resources.every((r) => perms.has(r));
}

/** Check if a role has ANY of the specified permissions */
export function hasAnyPermission(role: PlatformRole, resources: Resource[]): boolean {
  const perms = getPermissions(role);
  return resources.some((r) => perms.has(r));
}
