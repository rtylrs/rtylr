import { describe, it, expect } from 'vitest';
import { mapPOSUsers, buildStaffUpdate, type StaffUser } from './staffWorkflow';

describe('mapPOSUsers', () => {
  it('maps the users envelope (camelCase from the API layer)', () => {
    const out = mapPOSUsers({
      users: [
        {
          id: 5,
          name: 'Alice',
          role: 'manager',
          isActive: true,
          maxDiscountPct: 100,
          canVoid: true,
          canRefund: true,
          canOpenDrawer: true,
          canApplyDiscount: true,
        },
      ],
    });
    expect(out).toEqual([
      {
        id: 5,
        name: 'Alice',
        role: 'manager',
        isActive: true,
        maxDiscountPct: 100,
        canVoid: true,
        canRefund: true,
        canOpenDrawer: true,
        canApplyDiscount: true,
      },
    ]);
  });

  it('tolerates snake_case and a bare array', () => {
    const out = mapPOSUsers([
      { id: 9, name: 'Bob', role: 'cashier', is_active: false, can_void: false, max_discount_pct: 15 },
    ]);
    expect(out[0]).toMatchObject({ id: 9, name: 'Bob', role: 'cashier', isActive: false, maxDiscountPct: 15 });
  });

  it('falls back to cashier for an unknown role and drops invalid rows', () => {
    const out = mapPOSUsers({
      users: [
        { id: 0, name: 'Bad' },
        { id: 3, name: 'X', role: 'wizard' },
      ],
    });
    expect(out).toHaveLength(1);
    expect(out[0].role).toBe('cashier');
  });

  it('returns [] for unusable input', () => {
    expect(mapPOSUsers(null)).toEqual([]);
    expect(mapPOSUsers({})).toEqual([]);
  });
});

describe('buildStaffUpdate', () => {
  const user: StaffUser = {
    id: 5,
    name: 'Alice',
    role: 'cashier',
    isActive: true,
    maxDiscountPct: 0,
    canVoid: false,
    canRefund: false,
    canOpenDrawer: false,
    canApplyDiscount: false,
  };

  it('builds a snake_case update payload', () => {
    const payload = buildStaffUpdate(user, {
      role: 'manager',
      isActive: true,
      maxDiscountPct: 50,
      canVoid: true,
      canRefund: true,
      canOpenDrawer: true,
      canApplyDiscount: true,
    });
    expect(payload).toEqual({
      id: 5,
      name: 'Alice',
      role: 'manager',
      is_active: true,
      max_discount_pct: 50,
      can_void: true,
      can_refund: true,
      can_open_drawer: true,
      can_apply_discount: true,
    });
  });

  it('includes pin only when a new one is provided', () => {
    const withPin = buildStaffUpdate(user, {
      role: 'cashier',
      isActive: true,
      maxDiscountPct: 0,
      canVoid: false,
      canRefund: false,
      canOpenDrawer: false,
      canApplyDiscount: false,
      newPin: '4321',
    });
    expect(withPin.pin).toBe('4321');

    const blank = buildStaffUpdate(user, {
      role: 'cashier',
      isActive: true,
      maxDiscountPct: 0,
      canVoid: false,
      canRefund: false,
      canOpenDrawer: false,
      canApplyDiscount: false,
      newPin: '   ',
    });
    expect('pin' in blank).toBe(false);
  });
});
