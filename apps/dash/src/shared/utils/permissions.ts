import { usePermissionsStore } from '../stores/permissionsStore';
import { useUserStore } from '../stores/userStore';

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
  const userStore = useUserStore.getState();

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

// Hook to use permissions in components
export const usePermissions = () => {
  const permissionsStore = usePermissionsStore();
  const userStore = useUserStore();

  // Sync permissions with user role
  React.useEffect(() => {
    if (userStore.role) {
      permissionsStore.setRole(userStore.role as any);
    }
  }, [userStore.role]);

  return {
    hasPermission: permissionsStore.hasPermission,
    permissions: permissionsStore.getPermissions(),
    role: permissionsStore.role,
  };
};

// Need to import React for the hook
import React from 'react';
