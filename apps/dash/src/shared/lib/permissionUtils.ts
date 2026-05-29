/**
 * Permission utility helpers for the Dash app.
 * All functions are pure with no side-effects.
 */

export type DashPermission =
  | 'view_dashboard'
  | 'view_reports'
  | 'export_data'
  | 'manage_settings'
  | 'manage_users'
  | 'manage_billing'
  | 'delete_records'
  | 'impersonate_users';

export type DashUserRole = 'owner' | 'admin' | 'manager' | 'staff' | 'viewer';

const ROLE_PERMISSIONS: Record<DashUserRole, DashPermission[]> = {
  owner: [
    'view_dashboard',
    'view_reports',
    'export_data',
    'manage_settings',
    'manage_users',
    'manage_billing',
    'delete_records',
    'impersonate_users',
  ],
  admin: [
    'view_dashboard',
    'view_reports',
    'export_data',
    'manage_settings',
    'manage_users',
    'delete_records',
  ],
  manager: ['view_dashboard', 'view_reports', 'export_data', 'manage_settings'],
  staff: ['view_dashboard', 'view_reports'],
  viewer: ['view_dashboard'],
};

/** Return all permissions granted to a role. */
export function permissionsForRole(role: DashUserRole): DashPermission[] {
  return ROLE_PERMISSIONS[role] ?? [];
}

/** Return true when the role has the specified permission. */
export function hasPermission(role: DashUserRole, permission: DashPermission): boolean {
  return permissionsForRole(role).includes(permission);
}

/** Return true when the role has ALL of the specified permissions. */
export function hasAllPermissions(role: DashUserRole, permissions: DashPermission[]): boolean {
  const granted = permissionsForRole(role);
  return permissions.every((p) => granted.includes(p));
}

/** Return true when the role has AT LEAST ONE of the specified permissions. */
export function hasAnyPermission(role: DashUserRole, permissions: DashPermission[]): boolean {
  const granted = permissionsForRole(role);
  return permissions.some((p) => granted.includes(p));
}

/** Return the permissions that two roles share in common. */
export function sharedPermissions(a: DashUserRole, b: DashUserRole): DashPermission[] {
  const bSet = new Set(permissionsForRole(b));
  return permissionsForRole(a).filter((p) => bSet.has(p));
}

/** Return permissions that role `a` has but role `b` does not. */
export function extraPermissions(a: DashUserRole, b: DashUserRole): DashPermission[] {
  const bSet = new Set(permissionsForRole(b));
  return permissionsForRole(a).filter((p) => !bSet.has(p));
}
