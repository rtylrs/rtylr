import { describe, it, expect } from 'vitest';
import { mapPOSUsers, buildStaffUpdate, type StaffUser } from './staffWorkflow';
import { PIN_MAX_LENGTH, PIN_MIN_LENGTH } from '../login/pinPolicy';

describe('mapPOSUsers', () => {
  it('maps a decoded POSUser array (camelCase from the API layer)', () => {
    const out = mapPOSUsers([
      {
        id: '5',
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

  it('applies capability defaults when fields are absent', () => {
    const out = mapPOSUsers([
      { id: '9', name: 'Bob', role: 'cashier', isActive: false, canVoid: false, maxDiscountPct: 15 },
    ]);
    expect(out[0]).toMatchObject({ id: 9, name: 'Bob', role: 'cashier', isActive: false, maxDiscountPct: 15 });
  });

  it('falls back to cashier for an unknown role and drops invalid rows', () => {
    const out = mapPOSUsers([
      { id: '0', name: 'Bad' },
      { id: '3', name: 'X', role: 'wizard' },
    ]);
    expect(out).toHaveLength(1);
    expect(out[0].role).toBe('cashier');
  });

  it('returns [] for an empty array', () => {
    expect(mapPOSUsers([])).toEqual([]);
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

  const baseEdits = {
    role: 'cashier',
    isActive: true,
    maxDiscountPct: 0,
    canVoid: false,
    canRefund: false,
    canOpenDrawer: false,
    canApplyDiscount: false,
  } as const;

  it('builds the generated camelCase update payload', () => {
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
      id: '5',
      name: 'Alice',
      isActive: true,
    });
    // The rate is unknown (list response has no hourlyRate), so it must be
    // omitted rather than zeroed on every edit.
    expect('hourlyRate' in payload).toBe(false);
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
      newPin: '654321',
    });
    expect(withPin.pin).toBe('654321');

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

  it('accepts the server minimum and every length up to its maximum', () => {
    for (let length = PIN_MIN_LENGTH; length <= PIN_MAX_LENGTH; length += 1) {
      const payload = buildStaffUpdate(user, { ...baseEdits, newPin: '7'.repeat(length) });
      expect(payload.pin).toBe('7'.repeat(length));
    }
  });

  it('rejects a PIN the server would refuse, naming the accepted range', () => {
    expect(() => buildStaffUpdate(user, { ...baseEdits, newPin: '1234' })).toThrow(
      `PIN must be between ${PIN_MIN_LENGTH} and ${PIN_MAX_LENGTH} digits.`
    );
    expect(() => buildStaffUpdate(user, { ...baseEdits, newPin: '12345' })).toThrow();
    expect(() => buildStaffUpdate(user, { ...baseEdits, newPin: '1'.repeat(PIN_MAX_LENGTH + 1) })).toThrow();
    expect(() => buildStaffUpdate(user, { ...baseEdits, newPin: '12a456' })).toThrow();
  });
});
