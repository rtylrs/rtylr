import { describe, expect, it } from 'vitest';
import { createEffectiveAccess } from '../../../../../test/effectiveAccess';
import { canHrPermission, canViewCompensation } from './hrPermissions';

describe('canHrPermission', () => {
  it('allows owners and admins every action in a subscribed HR app', () => {
    for (const access of [
      createEffectiveAccess('hr', { isOwner: true }),
      createEffectiveAccess('hr', { isAdmin: true }),
    ]) {
      expect(canHrPermission('hr:payroll:execute', access)).toBe(true);
      expect(canHrPermission('hr:time_off:approve', access)).toBe(true);
    }
  });

  it('limits a member to explicitly assigned self-service permissions', () => {
    const access = createEffectiveAccess('hr', { permissions: ['hr:time_off:request'] });
    expect(canHrPermission('hr:time_off:request', access)).toBe(true);
    expect(canHrPermission('hr:employees:view', access)).toBe(false);
    expect(canHrPermission('hr:time_off:approve', access)).toBe(false);
    expect(canHrPermission('hr:payroll:view', access)).toBe(false);
  });

  it('denies permissions when effective access is missing', () => {
    expect(canHrPermission('hr:employees:view', null)).toBe(false);
    expect(canHrPermission('hr:payroll:view', undefined)).toBe(false);
  });

  it('limits compensation data to assigned sensitive permissions or admins', () => {
    const memberAccess = createEffectiveAccess('hr', {
      permissions: ['hr:employees:sensitive_view'],
    });
    expect(canViewCompensation(memberAccess)).toBe(true);
    expect(canViewCompensation(createEffectiveAccess('hr'))).toBe(false);
    expect(canViewCompensation(createEffectiveAccess('hr', { isAdmin: true }))).toBe(true);
    expect(canViewCompensation(null)).toBe(false);
  });
});
