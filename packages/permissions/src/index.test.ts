import { describe, it, expect } from 'vitest';
import {
  ROLE_TIERS,
  ROLE_PARENT,
  hasPOSRoleLevel,
  hasPermission,
  hasAllPermissions,
  hasAnyPermission,
  getPermissions,
  canAccessApp,
  getAccessibleApps,
  canPOSAction,
  can,
  canAll,
  canAny,
  canEffectiveAccess,
  canEffectiveAccessInScope,
} from './index';
import type { PlatformRole, PermissionContext } from './index';
import type { EffectiveAccess } from '@voxire/types';

// ── Role hierarchy ──────────────────────────────────────────────────

describe('ROLE_PARENT and ROLE_TIERS agree', () => {
  // ROLE_TIERS is otherwise only documentation. This makes it load-bearing: the
  // chain from owner down through ROLE_PARENT must be exactly the tier ladder,
  // so moving a lateral role into that chain fails here instead of silently
  // granting it everything below (which is what #795 was).
  it('the parent chain from owner is exactly ROLE_TIERS reversed', () => {
    const chain: PlatformRole[] = [];
    for (let r: PlatformRole | null = 'owner'; r !== null; r = ROLE_PARENT[r]) chain.push(r);
    expect(chain).toEqual([...ROLE_TIERS].reverse());
  });

  it('store_manager resolves to exactly staff, on purpose', () => {
    // Not a placeholder. A store manager's extra powers are location-scoped via
    // pos_member_locations and the PIN-verify capability flags, so none of them
    // is a platform-wide grant (#795). If this ever fails because someone added
    // permissions to store_manager, check that the intent really changed rather
    // than making the test agree with the code: granting globally what is meant
    // to be granted per location is the failure mode.
    expect([...getPermissions('store_manager')].sort()).toEqual(
      [...getPermissions('staff')].sort()
    );
  });

  it('every lateral role names a base outside the lateral set', () => {
    const laterals: PlatformRole[] = ['billing', 'menu_manager', 'store_manager'];
    for (const lateral of laterals) {
      expect(laterals).not.toContain(ROLE_PARENT[lateral]);
    }
  });

  it('billing does not reach the shop floor', () => {
    // The concrete #795 regression: billing held pos:drawer:open and hr:staff.
    const billing = getPermissions('billing');
    expect(billing.has('pos:drawer:open')).toBe(false);
    expect(billing.has('pos:sessions:close')).toBe(false);
    expect(billing.has('hr:staff')).toBe(false);
    expect(billing.has('app:pos')).toBe(false);
    // What it should keep.
    expect(billing.has('hub:billing')).toBe(true);
    expect(billing.has('finance:invoices')).toBe(true);
    expect(billing.has('hub:dashboard')).toBe(true);
  });

  it('staff keeps HR self-service but not the directory', () => {
    const staff = getPermissions('staff');
    expect(staff.has('hr:shifts')).toBe(true);
    expect(staff.has('hr:time_off')).toBe(true);
    expect(staff.has('hr:staff')).toBe(false);
    expect(getPermissions('operator').has('hr:staff')).toBe(true);
  });

  it('the tiers are unchanged by the restructure', () => {
    // Guards the regression this change could most easily have caused: operator
    // and above previously inherited app:menu, hub:design, hub:billing,
    // finance:invoices and finance:reports only because the lateral roles sat
    // below them in the flat list.
    for (const resource of [
      'app:menu',
      'hub:design',
      'hub:billing',
      'finance:invoices',
      'finance:reports',
    ] as const) {
      expect(getPermissions('operator').has(resource)).toBe(true);
      expect(getPermissions('admin').has(resource)).toBe(true);
      expect(getPermissions('owner').has(resource)).toBe(true);
    }
  });
});

describe('hasPOSRoleLevel', () => {
  it('manager is at least cashier', () => {
    expect(hasPOSRoleLevel('manager', 'cashier')).toBe(true);
  });

  it('cashier is NOT at least manager', () => {
    expect(hasPOSRoleLevel('cashier', 'manager')).toBe(false);
  });

  it('kitchen is at least kitchen', () => {
    expect(hasPOSRoleLevel('kitchen', 'kitchen')).toBe(true);
  });
});

// ── Permission matrix ───────────────────────────────────────────────

describe('getPermissions', () => {
  it('owner has all permissions', () => {
    const perms = getPermissions('owner');
    expect(perms.has('hub:dashboard')).toBe(true);
    expect(perms.has('app:pos')).toBe(true);
    expect(perms.has('pos:orders:void')).toBe(true);
    expect(perms.has('hr:payroll:run')).toBe(true);
    expect(perms.has('finance:reconcile')).toBe(true);
  });

  it('admin inherits operator permissions', () => {
    const perms = getPermissions('admin');
    expect(perms.has('app:erp')).toBe(true);
    expect(perms.has('erp:products')).toBe(true);
    expect(perms.has('crm:customers')).toBe(true);
  });

  it('staff has basic POS access', () => {
    const perms = getPermissions('staff');
    expect(perms.has('app:pos')).toBe(true);
    expect(perms.has('pos:orders')).toBe(true);
    expect(perms.has('pos:orders:create')).toBe(true);
  });

  it('staff does NOT have admin permissions', () => {
    const perms = getPermissions('staff');
    expect(perms.has('hub:settings')).toBe(false);
    expect(perms.has('pos:orders:void')).toBe(false);
    expect(perms.has('hr:payroll')).toBe(false);
  });

  it('viewer only has dashboard access', () => {
    const perms = getPermissions('viewer');
    expect(perms.has('hub:dashboard')).toBe(true);
    expect(perms.has('insights:dashboard')).toBe(true);
    expect(perms.has('app:pos')).toBe(false);
    expect(perms.has('app:erp')).toBe(false);
  });

  it('operator has CRM and ERP but NOT finance admin', () => {
    const perms = getPermissions('operator');
    expect(perms.has('app:crm')).toBe(true);
    expect(perms.has('app:erp')).toBe(true);
    expect(perms.has('finance:tax')).toBe(false);
    expect(perms.has('finance:reconcile')).toBe(false);
  });
});

describe('hasPermission', () => {
  it('returns true for valid permission', () => {
    expect(hasPermission('admin', 'hub:settings')).toBe(true);
  });

  it('returns false for invalid permission', () => {
    expect(hasPermission('staff', 'hub:settings')).toBe(false);
  });
});

describe('hasAllPermissions', () => {
  it('admin has all hub permissions', () => {
    expect(hasAllPermissions('admin', ['hub:settings', 'hub:addons', 'hub:developer'])).toBe(true);
  });

  it('staff does NOT have all hub permissions', () => {
    expect(hasAllPermissions('staff', ['hub:settings', 'hub:addons'])).toBe(false);
  });
});

describe('hasAnyPermission', () => {
  it('staff has at least one POS permission', () => {
    expect(hasAnyPermission('staff', ['pos:orders', 'hub:settings'])).toBe(true);
  });

  it('viewer has none of the admin permissions', () => {
    expect(hasAnyPermission('viewer', ['hub:settings', 'pos:orders:void'])).toBe(false);
  });
});

// ── App access ──────────────────────────────────────────────────────

describe('canAccessApp', () => {
  it('owner can access all apps', () => {
    expect(canAccessApp('owner', 'dash')).toBe(true);
    expect(canAccessApp('owner', 'pos')).toBe(true);
    expect(canAccessApp('owner', 'erp')).toBe(true);
    expect(canAccessApp('owner', 'finance')).toBe(true);
  });

  it('staff can access POS but not ERP', () => {
    expect(canAccessApp('staff', 'pos')).toBe(true);
    expect(canAccessApp('staff', 'erp')).toBe(false);
  });

  it('viewer cannot access POS', () => {
    expect(canAccessApp('viewer', 'pos')).toBe(false);
  });

  it('operator can access CRM', () => {
    expect(canAccessApp('operator', 'crm')).toBe(true);
  });

  it('billing role can access finance', () => {
    expect(canAccessApp('billing', 'finance')).toBe(true);
  });
});

describe('getAccessibleApps', () => {
  it('owner gets all apps', () => {
    const apps = getAccessibleApps('owner');
    expect(apps).toContain('dash');
    expect(apps).toContain('pos');
    expect(apps).toContain('erp');
    expect(apps).toContain('finance');
    expect(apps.length).toBe(10); // All app IDs
  });

  it('viewer gets minimal apps', () => {
    const apps = getAccessibleApps('viewer');
    expect(apps).toContain('dash');
    expect(apps).not.toContain('pos');
  });

  it('staff gets operational apps', () => {
    const apps = getAccessibleApps('staff');
    expect(apps).toContain('pos');
    expect(apps).toContain('flow');
    expect(apps).toContain('hr');
    expect(apps).not.toContain('erp');
    expect(apps).not.toContain('finance');
  });
});

// ── POS action guards ───────────────────────────────────────────────

describe('canPOSAction', () => {
  it('manager can void orders', () => {
    expect(canPOSAction('manager', 'void_order')).toBe(true);
  });

  it('cashier cannot void orders', () => {
    expect(canPOSAction('cashier', 'void_order')).toBe(false);
  });

  it('cashier can apply discounts', () => {
    expect(canPOSAction('cashier', 'apply_discount')).toBe(true);
  });

  it('server can open drawer', () => {
    expect(canPOSAction('server', 'open_drawer')).toBe(true);
  });

  it('kitchen cannot open drawer', () => {
    expect(canPOSAction('kitchen', 'open_drawer')).toBe(false);
  });

  it('manager can manage employees', () => {
    expect(canPOSAction('manager', 'manage_employees')).toBe(true);
  });

  it('cashier cannot manage employees', () => {
    expect(canPOSAction('cashier', 'manage_employees')).toBe(false);
  });
});

// ── Universal permission check ──────────────────────────────────────

describe('can', () => {
  it('owner bypasses everything', () => {
    const ctx: PermissionContext = { platformRole: 'owner' };
    expect(can(ctx, 'hr:payroll:run')).toBe(true);
    expect(can(ctx, 'finance:reconcile')).toBe(true);
  });

  it('staff with assigned apps can access those apps', () => {
    const ctx: PermissionContext = {
      platformRole: 'staff',
      assignedApps: ['erp', 'crm'],
    };
    expect(can(ctx, 'app:erp')).toBe(true);
    expect(can(ctx, 'app:crm')).toBe(true);
    expect(can(ctx, 'app:finance')).toBe(false);
  });

  it('falls back to role permissions when no app assignment', () => {
    const ctx: PermissionContext = { platformRole: 'admin' };
    expect(can(ctx, 'hub:settings')).toBe(true);
    expect(can(ctx, 'pos:orders:void')).toBe(true);
  });
});

describe('canAll', () => {
  it('checks all permissions', () => {
    const ctx: PermissionContext = { platformRole: 'admin' };
    expect(canAll(ctx, ['hub:settings', 'hub:addons'])).toBe(true);
    expect(canAll(ctx, ['hub:settings', 'hr:payroll:run'])).toBe(false);
  });
});

describe('canAny', () => {
  it('checks any permission', () => {
    const ctx: PermissionContext = { platformRole: 'staff' };
    expect(canAny(ctx, ['app:pos', 'hub:settings'])).toBe(true);
    expect(canAny(ctx, ['hub:settings', 'finance:reconcile'])).toBe(false);
  });
});

describe('canEffectiveAccess', () => {
  const access: EffectiveAccess = {
    businessId: 'business-1',
    dashRole: 'manager',
    dashRoleId: 'dash-manager',
    dashPermissions: ['hub:team:view'],
    subscribedApps: ['flow', 'finance'],
    licenses: [],
    apps: [
      {
        appId: 'flow',
        licensed: true,
        hasAccess: true,
        roleIds: ['flow-manager'],
        primaryRoleId: 'flow-manager',
        displayRole: 'Manager',
        permissions: ['flow:tasks:view', 'flow:tasks:edit'],
        scopes: [{ scopeType: 'project', scopeId: 'project-1' }],
        sources: ['explicit'],
        permissionSources: {
          'flow:tasks:view': ['flow-manager'],
          'flow:tasks:edit': ['flow-manager'],
        },
      },
      {
        appId: 'finance',
        licensed: false,
        hasAccess: false,
        roleIds: ['finance-viewer'],
        primaryRoleId: 'finance-viewer',
        displayRole: 'Viewer',
        permissions: ['finance:reports:view'],
        scopes: [],
        sources: ['explicit'],
        permissionSources: { 'finance:reports:view': ['finance-viewer'] },
      },
    ],
    isOwner: false,
    isAdmin: false,
    authorizationVersion: 2,
  };

  it('requires both a license and a resolved permission', () => {
    expect(canEffectiveAccess(access, 'flow:tasks:edit')).toBe(true);
    expect(canEffectiveAccess(access, 'finance:reports:view')).toBe(false);
  });

  it('does not turn an app assignment into child permissions', () => {
    expect(canEffectiveAccess(access, 'app:flow')).toBe(true);
    expect(canEffectiveAccess(access, 'flow:projects:delete')).toBe(false);
  });

  it('requires the matching scope for row-level checks', () => {
    expect(
      canEffectiveAccessInScope(access, 'flow:tasks:edit', {
        scopeType: 'project',
        scopeId: 'project-1',
      })
    ).toBe(true);
    expect(
      canEffectiveAccessInScope(access, 'flow:tasks:edit', {
        scopeType: 'project',
        scopeId: 'project-2',
      })
    ).toBe(false);
  });
});
