import { describe, expect, it } from 'vitest';
import { canHrPermission, canViewCompensation } from './hrPermissions';

describe('canHrPermission', () => {
  it('allows owners to use every HR action', () => {
    expect(canHrPermission('owner', 'hr:payroll:run')).toBe(true);
    expect(canHrPermission('owner', 'hr:time_off:approve')).toBe(true);
  });

  it('keeps management actions away from staff', () => {
    expect(canHrPermission('staff', 'hr:staff:manage')).toBe(false);
    expect(canHrPermission('staff', 'hr:time_off:approve')).toBe(false);
  });

  it('gives staff HR self-service but not the employee directory', () => {
    // hr:staff gates /staff, /staff/$id, /employees and /employees/$id, which
    // is every employee's record rather than the person's own. Staff held it
    // until #795; they keep the self-service half.
    expect(canHrPermission('staff', 'hr:staff')).toBe(false);
    expect(canHrPermission('staff', 'hr:shifts')).toBe(true);
    expect(canHrPermission('staff', 'hr:time_off')).toBe(true);
    // The directory starts at operator.
    expect(canHrPermission('operator', 'hr:staff')).toBe(true);
  });

  it('does not let a billing role read HR', () => {
    // billing inherited staff's grants through the old flat role list, so it
    // held hr:staff, hr:shifts and hr:time_off despite being a finance role.
    expect(canHrPermission('billing', 'hr:staff')).toBe(false);
    expect(canHrPermission('billing', 'hr:shifts')).toBe(false);
    expect(canHrPermission('billing', 'hr:time_off')).toBe(false);
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
