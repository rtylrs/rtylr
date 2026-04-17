import { describe, it, expect, beforeEach } from 'vitest';
import { usePermissionsStore, type UserRole } from './permissionsStore';

const { getState } = usePermissionsStore;

beforeEach(() => {
  usePermissionsStore.setState({ role: null, permissions: null });
});

describe('permissionsStore', () => {
  it('has correct initial state', () => {
    expect(getState().role).toBeNull();
    expect(getState().permissions).toBeNull();
  });

  it('setRole sets owner with full permissions', () => {
    getState().setRole('owner');
    expect(getState().role).toBe('owner');
    expect(getState().permissions?.canEditProducts).toBe(true);
    expect(getState().permissions?.canManageBilling).toBe(true);
    expect(getState().permissions?.canManageUsers).toBe(true);
  });

  it('setRole sets admin with full permissions', () => {
    getState().setRole('admin');
    expect(getState().role).toBe('admin');
    expect(getState().permissions?.canEditProducts).toBe(true);
    expect(getState().permissions?.canManageBilling).toBe(true);
  });

  it('setRole sets manager with limited permissions', () => {
    getState().setRole('manager');
    expect(getState().role).toBe('manager');
    expect(getState().permissions?.canEditProducts).toBe(true);
    expect(getState().permissions?.canManageSettings).toBe(false);
    expect(getState().permissions?.canManageUsers).toBe(false);
    expect(getState().permissions?.canManageBilling).toBe(false);
  });

  it('setRole sets staff with minimal permissions', () => {
    getState().setRole('staff');
    expect(getState().permissions?.canManageOrders).toBe(true);
    expect(getState().permissions?.canEditProducts).toBe(false);
    expect(getState().permissions?.canViewAnalytics).toBe(false);
    expect(getState().permissions?.canManageMenu).toBe(false);
  });

  it('setRole sets viewer with read-only permissions', () => {
    getState().setRole('viewer');
    expect(getState().permissions?.canViewAnalytics).toBe(true);
    expect(getState().permissions?.canViewReports).toBe(true);
    expect(getState().permissions?.canEditProducts).toBe(false);
    expect(getState().permissions?.canManageOrders).toBe(false);
  });

  it('hasPermission returns true for allowed permission', () => {
    getState().setRole('owner');
    expect(getState().hasPermission('canEditProducts')).toBe(true);
  });

  it('hasPermission returns false for denied permission', () => {
    getState().setRole('staff');
    expect(getState().hasPermission('canEditProducts')).toBe(false);
  });

  it('hasPermission returns false when no role set', () => {
    expect(getState().hasPermission('canEditProducts')).toBe(false);
  });

  it('getPermissions returns null when no role set', () => {
    expect(getState().getPermissions()).toBeNull();
  });

  it('getPermissions returns permissions when role is set', () => {
    getState().setRole('owner');
    const perms = getState().getPermissions();
    expect(perms).not.toBeNull();
    expect(perms?.canEditProducts).toBe(true);
  });

  it('changing role updates permissions', () => {
    getState().setRole('owner');
    expect(getState().hasPermission('canManageBilling')).toBe(true);

    getState().setRole('staff');
    expect(getState().hasPermission('canManageBilling')).toBe(false);
  });
});
