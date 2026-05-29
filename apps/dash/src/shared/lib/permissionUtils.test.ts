import { describe, it, expect } from 'vitest';
import {
  permissionsForRole,
  hasPermission,
  hasAllPermissions,
  hasAnyPermission,
  sharedPermissions,
  extraPermissions,
} from './permissionUtils';

describe('permissionsForRole', () => {
  it('returns all permissions for owner', () => {
    const perms = permissionsForRole('owner');
    expect(perms).toContain('impersonate_users');
    expect(perms).toContain('manage_billing');
    expect(perms.length).toBeGreaterThan(5);
  });

  it('returns only view_dashboard for viewer', () => {
    expect(permissionsForRole('viewer')).toEqual(['view_dashboard']);
  });
});

describe('hasPermission', () => {
  it('returns true for a permission the role has', () => {
    expect(hasPermission('admin', 'manage_users')).toBe(true);
  });

  it('returns false for a permission the role lacks', () => {
    expect(hasPermission('viewer', 'export_data')).toBe(false);
  });

  it('owner has manage_billing', () => {
    expect(hasPermission('owner', 'manage_billing')).toBe(true);
  });

  it('staff does not have manage_settings', () => {
    expect(hasPermission('staff', 'manage_settings')).toBe(false);
  });
});

describe('hasAllPermissions', () => {
  it('returns true when role has all specified permissions', () => {
    expect(hasAllPermissions('manager', ['view_dashboard', 'view_reports', 'export_data'])).toBe(
      true
    );
  });

  it('returns false when role is missing at least one permission', () => {
    expect(hasAllPermissions('manager', ['view_dashboard', 'manage_billing'])).toBe(false);
  });

  it('returns true for an empty permission list', () => {
    expect(hasAllPermissions('viewer', [])).toBe(true);
  });
});

describe('hasAnyPermission', () => {
  it('returns true when role has at least one permission', () => {
    expect(hasAnyPermission('staff', ['export_data', 'view_reports'])).toBe(true);
  });

  it('returns false when role has none of the permissions', () => {
    expect(hasAnyPermission('viewer', ['manage_billing', 'delete_records'])).toBe(false);
  });

  it('returns false for an empty permission list', () => {
    expect(hasAnyPermission('owner', [])).toBe(false);
  });
});

describe('sharedPermissions', () => {
  it('returns permissions common to both roles', () => {
    const shared = sharedPermissions('admin', 'manager');
    expect(shared).toContain('view_dashboard');
    expect(shared).toContain('view_reports');
    expect(shared).toContain('export_data');
  });

  it('does not include permissions unique to one role', () => {
    const shared = sharedPermissions('admin', 'viewer');
    expect(shared).toEqual(['view_dashboard']);
  });
});

describe('extraPermissions', () => {
  it('returns permissions role a has that b does not', () => {
    const extras = extraPermissions('admin', 'manager');
    expect(extras).toContain('manage_users');
    expect(extras).toContain('delete_records');
    expect(extras).not.toContain('view_dashboard');
  });

  it('returns empty array when a and b have the same permissions', () => {
    expect(extraPermissions('viewer', 'viewer')).toEqual([]);
  });
});
