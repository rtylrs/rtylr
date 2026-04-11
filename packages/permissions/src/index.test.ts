import { describe, it, expect } from 'vitest';
import {
  hasRoleLevel,
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
} from './index';
import type { PlatformRole, POSRole, Resource, PermissionContext } from './index';

// ── Role hierarchy ──────────────────────────────────────────────────

describe('hasRoleLevel', () => {
  it('owner is at least owner', () => {
    expect(hasRoleLevel('owner', 'owner')).toBe(true);
  });

  it('owner is at least viewer', () => {
    expect(hasRoleLevel('owner', 'viewer')).toBe(true);
  });

  it('viewer is NOT at least admin', () => {
    expect(hasRoleLevel('viewer', 'admin')).toBe(false);
  });

  it('admin is at least operator', () => {
    expect(hasRoleLevel('admin', 'operator')).toBe(true);
  });

  it('staff is NOT at least operator', () => {
    expect(hasRoleLevel('staff', 'operator')).toBe(false);
  });

  it('operator is at least staff', () => {
    expect(hasRoleLevel('operator', 'staff')).toBe(true);
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
    expect(canAccessApp('owner', 'hub')).toBe(true);
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

  it('operator can access CRM and store', () => {
    expect(canAccessApp('operator', 'crm')).toBe(true);
    expect(canAccessApp('operator', 'store')).toBe(true);
  });

  it('billing role can access finance', () => {
    expect(canAccessApp('billing', 'finance')).toBe(true);
  });
});

describe('getAccessibleApps', () => {
  it('owner gets all apps', () => {
    const apps = getAccessibleApps('owner');
    expect(apps).toContain('hub');
    expect(apps).toContain('pos');
    expect(apps).toContain('erp');
    expect(apps).toContain('finance');
    expect(apps.length).toBe(12); // All app IDs
  });

  it('viewer gets minimal apps', () => {
    const apps = getAccessibleApps('viewer');
    expect(apps).toContain('hub');
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
