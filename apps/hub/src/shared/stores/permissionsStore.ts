import { create } from 'zustand';
import { persist } from 'zustand/middleware';

export type UserRole = 'owner' | 'admin' | 'manager' | 'staff' | 'viewer';

interface Permission {
  canEditProducts: boolean;
  canViewAnalytics: boolean;
  canManageSettings: boolean;
  canManageUsers: boolean;
  canViewReports: boolean;
  canManageMenu: boolean;
  canManageOrders: boolean;
  canManageBilling: boolean;
}

const rolePermissions: Record<UserRole, Permission> = {
  owner: {
    canEditProducts: true,
    canViewAnalytics: true,
    canManageSettings: true,
    canManageUsers: true,
    canViewReports: true,
    canManageMenu: true,
    canManageOrders: true,
    canManageBilling: true,
  },
  admin: {
    canEditProducts: true,
    canViewAnalytics: true,
    canManageSettings: true,
    canManageUsers: true,
    canViewReports: true,
    canManageMenu: true,
    canManageOrders: true,
    canManageBilling: true,
  },
  manager: {
    canEditProducts: true,
    canViewAnalytics: true,
    canManageSettings: false,
    canManageUsers: false,
    canViewReports: true,
    canManageMenu: true,
    canManageOrders: true,
    canManageBilling: false,
  },
  staff: {
    canEditProducts: false,
    canViewAnalytics: false,
    canManageSettings: false,
    canManageUsers: false,
    canViewReports: false,
    canManageMenu: false,
    canManageOrders: true,
    canManageBilling: false,
  },
  viewer: {
    canEditProducts: false,
    canViewAnalytics: true,
    canManageSettings: false,
    canManageUsers: false,
    canViewReports: true,
    canManageMenu: false,
    canManageOrders: false,
    canManageBilling: false,
  },
};

interface PermissionsState {
  role: UserRole | null;
  permissions: Permission | null;
}

interface PermissionsActions {
  setRole: (role: UserRole) => void;
  getPermissions: () => Permission | null;
  hasPermission: (permission: keyof Permission) => boolean;
}

const initialState: PermissionsState = {
  role: null,
  permissions: null,
};

export const usePermissionsStore = create<PermissionsState & PermissionsActions>()(
  persist(
    (set, get) => ({
      ...initialState,

      setRole: (role) => {
        const permissions = rolePermissions[role];
        set({ role, permissions });
      },

      getPermissions: () => {
        const state = get();
        return state.permissions;
      },

      hasPermission: (permission) => {
        const state = get();
        if (!state.permissions) return false;
        return state.permissions[permission];
      },
    }),
    {
      name: 'permissions-store',
      partialize: (state) => ({
        role: state.role,
        permissions: state.permissions,
      }),
    }
  )
);

