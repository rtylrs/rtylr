import { describe, expect, it } from 'vitest';
import { createEffectiveAccess } from '../../../../../test/effectiveAccess';
import { canAccessCrmPath, getCrmHomePath } from './navigation';

const member = (permissions: string[] = []) => ({
  effectiveAccess: createEffectiveAccess('crm', { permissions }),
});

describe('CRM navigation inside Portal', () => {
  it('lands a campaigns-only member on campaigns instead of the customer dashboard', () => {
    const user = member(['crm:campaigns:view']);
    expect(getCrmHomePath(user)).toBe('/crm/campaigns');
    expect(canAccessCrmPath(user, '/crm/customers/1')).toBe(false);
    expect(canAccessCrmPath(user, '/crm/campaigns')).toBe(true);
    expect(canAccessCrmPath(user, '/crm/campaigns/email')).toBe(false);
  });
  it('separates customer reading, creation, import, and merge access', () => {
    const user = member(['crm:customers:view', 'crm:customers:create']);
    expect(getCrmHomePath(user)).toBe('/crm');
    expect(canAccessCrmPath(user, '/crm/customers/1')).toBe(true);
    expect(canAccessCrmPath(user, '/crm/customers/import')).toBe(true);
    expect(canAccessCrmPath(user, '/crm/customers/merge')).toBe(false);
  });
  it('denies CRM deep links when its license or app grant is absent', () => {
    expect(
      canAccessCrmPath(
        { effectiveAccess: createEffectiveAccess('crm', { licensed: false }) },
        '/crm'
      )
    ).toBe(false);
    expect(canAccessCrmPath(null, '/settings/crm')).toBe(false);
    expect(
      canAccessCrmPath({ effectiveAccess: createEffectiveAccess('flow') }, '/crm/customers')
    ).toBe(false);
  });
});
