// Roles
export { ROLE_HIERARCHY, POS_ROLE_HIERARCHY, hasRoleLevel, hasPOSRoleLevel } from './roles';
export type { PlatformRole, POSRole } from './roles';

// Capabilities
export type { Resource, Action } from './capabilities';

// Permission matrix
export { getPermissions, hasPermission, hasAllPermissions, hasAnyPermission } from './matrix';

// Guards
export { canAccessApp, getAccessibleApps, canPOSAction, can, canAll, canAny } from './guards';
export type { AppId, PermissionContext } from './guards';
