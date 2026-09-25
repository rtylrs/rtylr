// Roles
export { ROLE_TIERS, ROLE_PARENT, POS_ROLE_HIERARCHY, hasPOSRoleLevel } from './roles';
export type { PlatformRole, POSRole } from './roles';

// Capabilities
export type { Resource, Action } from './capabilities';

// Permission matrix
export { getPermissions, hasPermission, hasAllPermissions, hasAnyPermission } from './matrix';

// Guards
export {
  canAccessApp,
  getAccessibleApps,
  canPOSAction,
  can,
  canAll,
  canAny,
  canEffectiveAccess,
  canEffectiveAccessInScope,
  canonicalizePermission,
} from './guards';
export type { AppId, PermissionContext } from './guards';
