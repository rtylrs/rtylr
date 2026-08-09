/** Permission and role check helpers for the auth app. */

export type Role = string;

export interface PermissionConfig {
  roles: Role[];
  permissions: string[];
}

/**
 * Returns true if the given role is included in the allowed roles list.
 */
export function hasRole(userRole: Role, allowedRoles: Role[]): boolean {
  return allowedRoles.includes(userRole);
}

/**
 * Returns true if the user has all required permissions.
 */
export function hasPermissions(userPerms: string[], required: string[]): boolean {
  return required.every((p) => userPerms.includes(p));
}

/**
 * Returns true if the user has at least one of the given permissions.
 */
export function hasAnyPermission(userPerms: string[], anyOf: string[]): boolean {
  return anyOf.some((p) => userPerms.includes(p));
}

/**
 * Returns true if the user satisfies the role check and the permission check.
 */
export function canAccess(userRole: Role, userPerms: string[], config: PermissionConfig): boolean {
  return hasRole(userRole, config.roles) && hasPermissions(userPerms, config.permissions);
}

/**
 * Merges two permission lists, deduplicating entries.
 */
export function mergePermissions(a: string[], b: string[]): string[] {
  return Array.from(new Set([...a, ...b]));
}
