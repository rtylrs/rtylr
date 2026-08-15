import { describe, expect, it } from 'vitest';
import { canHrPermission, canViewCompensation } from './hrPermissions';

describe('canHrPermission', () => {
  it('allows owners to use every HR action', () => {
    expect(canHrPermission('owner', 'hr:payroll:run')).toBe(true);
    expect(canHrPermission('owner', 'hr:time_off:approve')).toBe(true);
  });

  it('keeps management actions away from staff', () => {
    expect(canHrPermission('staff', 'hr:staff')).toBe(true);
    expect(canHrPermission('staff', 'hr:staff:manage')).toBe(false);
    expect(canHrPermission('staff', 'hr:time_off:approve')).toBe(false);
  });

  it('allows an operator to manage shifts but not payroll', () => {
    expect(canHrPermission('operator', 'hr:shifts:manage')).toBe(true);
    expect(canHrPermission('operator', 'hr:payroll')).toBe(false);
    expect(canHrPermission('operator', 'hr:payroll:run')).toBe(false);
  });

  it('denies a missing role', () => {
    expect(canHrPermission(null, 'hr:staff')).toBe(false);
    expect(canHrPermission(undefined, 'hr:payroll')).toBe(false);
  });

  it('limits compensation visibility to staff or payroll managers', () => {
    expect(canViewCompensation('admin')).toBe(true);
    expect(canViewCompensation('owner')).toBe(true);
    expect(canViewCompensation('operator')).toBe(false);
    expect(canViewCompensation('staff')).toBe(false);
    expect(canViewCompensation(null)).toBe(false);
  });
});
