import { usePermissionsStore } from '../stores/permissionsStore';

export type PermissionKey =
  | 'canEditProducts'
  | 'canViewAnalytics'
  | 'canManageSettings'
  | 'canManageUsers'
  | 'canViewReports'
  | 'canManageMenu'
  | 'canManageOrders'
  | 'canManageBilling';

export const hasPermission = (permission: PermissionKey): boolean => {
  const permissionsStore = usePermissionsStore.getState();
  return permissionsStore.hasPermission(permission);
};

export const canAccessRoute = (route: string): boolean => {
  const permissionsStore = usePermissionsStore.getState();

  // Define route-permission mappings
  const routePermissions: Record<string, PermissionKey[]> = {
    '/settings': ['canManageSettings'],
    '/app/business/analytics': ['canViewAnalytics'],
    '/app/business/menus': ['canManageMenu'],
    '/app/business/menu': ['canManageMenu'],
    '/app/business/billing': ['canManageBilling'],
  };

  const requiredPermissions = routePermissions[route];
  if (!requiredPermissions) return true; // Route doesn't require specific permissions

  return requiredPermissions.some((permission) => permissionsStore.hasPermission(permission));
};

export const PERMISSIONS = {
  EDIT_PRODUCTS: 'canEditProducts' as PermissionKey,
  VIEW_ANALYTICS: 'canViewAnalytics' as PermissionKey,
  MANAGE_SETTINGS: 'canManageSettings' as PermissionKey,
  MANAGE_USERS: 'canManageUsers' as PermissionKey,
  VIEW_REPORTS: 'canViewReports' as PermissionKey,
  MANAGE_MENU: 'canManageMenu' as PermissionKey,
  MANAGE_ORDERS: 'canManageOrders' as PermissionKey,
  MANAGE_BILLING: 'canManageBilling' as PermissionKey,
};
