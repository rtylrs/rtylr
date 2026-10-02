import { describe, expect, it } from 'vitest';
import { createEffectiveAccess } from '../../../../../../../test/effectiveAccess';
import { canCrmPermission } from './crmPermissions';

describe('canCrmPermission', () => {
  it('allows owners and admins every action in a subscribed CRM app', () => {
    for (const access of [
      createEffectiveAccess('crm', { isOwner: true }),
      createEffectiveAccess('crm', { isAdmin: true }),
    ]) {
      expect(canCrmPermission('crm:campaigns:send', access)).toBe(true);
      expect(canCrmPermission('crm:loyalty:manage', access)).toBe(true);
      expect(canCrmPermission('crm:automations:manage', access)).toBe(true);
      expect(canCrmPermission('hub:team:invite', access)).toBe(true);
    }
  });

  it('grants a member only the CRM permissions assigned to their profiles', () => {
    const access = createEffectiveAccess('crm', {
      permissions: ['crm:customers:view', 'crm:customers:edit'],
    });
    expect(canCrmPermission('crm:customers:view', access)).toBe(true);
    expect(canCrmPermission('crm:customers:edit', access)).toBe(true);
    expect(canCrmPermission('crm:customers:delete', access)).toBe(false);
    expect(canCrmPermission('crm:campaigns:send', access)).toBe(false);
    expect(canCrmPermission('hub:team:invite', access)).toBe(false);
  });

  it('denies permissions when effective access is missing', () => {
    expect(canCrmPermission('crm:customers:view', null)).toBe(false);
    expect(canCrmPermission('crm:campaigns:send', undefined)).toBe(false);
  });
});
