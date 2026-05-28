import type { PlatformRole } from './roles';
import type { POSRole } from './roles';
import type { Resource } from './capabilities';
import { hasPermission, hasAnyPermission } from './matrix';
import { hasPOSRoleLevel } from './roles';

// ── App access guard ────────────────────────────────────────────────

export type AppId =
  | 'dash'
  | 'pos'
  | 'erp'
  | 'hr'
  | 'crm'
  | 'finance'
  | 'flow'
  | 'insights'
  | 'menu';

const APP_RESOURCE_MAP: Record<AppId, Resource> = {
  dash: 'hub:dashboard',
  pos: 'app:pos',
  erp: 'app:erp',
  hr: 'app:hr',
  crm: 'app:crm',
  finance: 'finance:invoices',
  flow: 'app:flow',
  insights: 'app:insights',
  menu: 'app:menu',
};

/** Check if a platform role can access a specific app */
export function canAccessApp(role: PlatformRole, appId: AppId): boolean {
  // Owner and admin can access everything
  if (role === 'owner' || role === 'admin') return true;
  const resource = APP_RESOURCE_MAP[appId];
  if (!resource) return false;
  return hasPermission(role, resource);
}

/** Get list of apps a role can access */
export function getAccessibleApps(role: PlatformRole): AppId[] {
  return (Object.keys(APP_RESOURCE_MAP) as AppId[]).filter((appId) => canAccessApp(role, appId));
}

// ── POS action guards ───────────────────────────────────────────────

/** Check if a POS user can perform a specific action */
export function canPOSAction(
  posRole: POSRole,
  action:
    | 'void_order'
    | 'refund_order'
    | 'apply_discount'
    | 'open_drawer'
    | 'manage_employees'
    | 'manage_terminals'
    | 'view_reports'
    | 'cash_drop'
    | 'cash_pickup'
): boolean {
  switch (action) {
    case 'void_order':
    case 'refund_order':
    case 'manage_employees':
    case 'manage_terminals':
    case 'view_reports':
    case 'cash_pickup':
      return posRole === 'manager';
    case 'apply_discount':
    case 'cash_drop':
      return posRole === 'manager' || posRole === 'cashier';
    case 'open_drawer':
      return posRole === 'manager' || posRole === 'cashier' || posRole === 'server';
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
}

/** Universal permission check - combines platform role, POS role, and app assignments */
export function can(ctx: PermissionContext, resource: Resource): boolean {
  // Owner bypasses everything
  if (ctx.platformRole === 'owner') return true;

  // Check platform permission
  if (hasPermission(ctx.platformRole, resource)) return true;

  // If they have explicit app assignments, check those
  if (ctx.assignedApps && ctx.assignedApps.length > 0) {
    const appId = resource.split(':')[0] as string;
    if (appId === 'app') {
      const app = resource.split(':')[1];
      return ctx.assignedApps.includes(app);
    }
  }

  return false;
}

/** Check multiple permissions at once */
export function canAll(ctx: PermissionContext, resources: Resource[]): boolean {
  return resources.every((r) => can(ctx, r));
}

export function canAny(ctx: PermissionContext, resources: Resource[]): boolean {
  return resources.some((r) => can(ctx, r));
}
