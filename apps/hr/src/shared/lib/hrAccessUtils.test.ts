import { describe, it, expect } from 'vitest';
import {
  hrRoleLabel,
  canManagePayroll,
  canViewAllEmployees,
  canApproveLeave,
  canEditSettings,
  canExportData,
  hrRoleRank,
  hasHigherRole,
} from './hrAccessUtils';

describe('hrRoleLabel', () => {
  it('returns Administrator for admin', () => {
    expect(hrRoleLabel('admin')).toBe('Administrator');
  });
  it('returns HR Staff for hr_staff', () => {
    expect(hrRoleLabel('hr_staff')).toBe('HR Staff');
  });
  it('returns Viewer for viewer', () => {
    expect(hrRoleLabel('viewer')).toBe('Viewer');
  });
});

describe('canManagePayroll', () => {
  it('allows admin', () => {
    expect(canManagePayroll('admin')).toBe(true);
  });
  it('allows hr_staff', () => {
    expect(canManagePayroll('hr_staff')).toBe(true);
  });
  it('denies manager', () => {
    expect(canManagePayroll('manager')).toBe(false);
  });
  it('denies employee', () => {
    expect(canManagePayroll('employee')).toBe(false);
  });
});

describe('canViewAllEmployees', () => {
  it('allows admin', () => {
    expect(canViewAllEmployees('admin')).toBe(true);
  });
  it('allows manager', () => {
    expect(canViewAllEmployees('manager')).toBe(true);
  });
  it('denies employee', () => {
    expect(canViewAllEmployees('employee')).toBe(false);
  });
});

describe('canApproveLeave', () => {
  it('allows admin', () => {
    expect(canApproveLeave('admin')).toBe(true);
  });
  it('allows manager', () => {
    expect(canApproveLeave('manager')).toBe(true);
  });
  it('denies viewer', () => {
    expect(canApproveLeave('viewer')).toBe(false);
  });
});

describe('canEditSettings', () => {
  it('allows admin only', () => {
    expect(canEditSettings('admin')).toBe(true);
    expect(canEditSettings('manager')).toBe(false);
    expect(canEditSettings('hr_staff')).toBe(false);
  });
});

describe('canExportData', () => {
  it('allows admin', () => {
    expect(canExportData('admin')).toBe(true);
  });
  it('allows hr_staff', () => {
    expect(canExportData('hr_staff')).toBe(true);
  });
  it('denies viewer', () => {
    expect(canExportData('viewer')).toBe(false);
  });
});

describe('hrRoleRank', () => {
  it('admin has highest rank', () => {
    expect(hrRoleRank('admin')).toBe(5);
  });
  it('viewer has lowest rank', () => {
    expect(hrRoleRank('viewer')).toBe(1);
  });
  it('manager ranks above hr_staff', () => {
    expect(hrRoleRank('manager')).toBeGreaterThan(hrRoleRank('hr_staff'));
  });
});

describe('hasHigherRole', () => {
  it('admin is higher than employee', () => {
    expect(hasHigherRole('admin', 'employee')).toBe(true);
  });
  it('employee is not higher than manager', () => {
    expect(hasHigherRole('employee', 'manager')).toBe(false);
  });
  it('same role is not higher', () => {
    expect(hasHigherRole('admin', 'admin')).toBe(false);
  });
});
