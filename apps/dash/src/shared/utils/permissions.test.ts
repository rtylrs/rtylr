import { describe, it, expect } from 'vitest';
import { PERMISSIONS, canAccessRoute } from './permissions';

describe('PERMISSIONS constants', () => {
  it('exposes EDIT_PRODUCTS', () => {
    expect(PERMISSIONS.EDIT_PRODUCTS).toBe('canEditProducts');
  });
  it('exposes VIEW_ANALYTICS', () => {
    expect(PERMISSIONS.VIEW_ANALYTICS).toBe('canViewAnalytics');
  });
  it('exposes MANAGE_SETTINGS', () => {
    expect(PERMISSIONS.MANAGE_SETTINGS).toBe('canManageSettings');
  });
  it('exposes MANAGE_MENU', () => {
    expect(PERMISSIONS.MANAGE_MENU).toBe('canManageMenu');
  });
  it('exposes MANAGE_ORDERS', () => {
    expect(PERMISSIONS.MANAGE_ORDERS).toBe('canManageOrders');
  });
});

describe('canAccessRoute', () => {
  it('returns true for routes with no permission requirements', () => {
    expect(canAccessRoute('/app/business/team')).toBe(true);
  });
  it('returns true for unrecognized routes', () => {
    expect(canAccessRoute('/some/random/route')).toBe(true);
  });
});
