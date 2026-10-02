import { describe, expect, it } from 'vitest';
import { createEffectiveAccess } from '../../../../../../../test/effectiveAccess';
import { canInsightsPermission } from './insightsPermissions';

describe('canInsightsPermission', () => {
  it('allows owners and admins every action in a subscribed Insights app', () => {
    for (const access of [
      createEffectiveAccess('insights', { isOwner: true }),
      createEffectiveAccess('insights', { isAdmin: true }),
    ]) {
      expect(canInsightsPermission('insights:reports:export', access)).toBe(true);
    }
  });

  it('lets a member view reports without granting export', () => {
    const access = createEffectiveAccess('insights', {
      permissions: ['insights:reports:create'],
    });
    expect(canInsightsPermission('insights:reports:create', access)).toBe(true);
    expect(canInsightsPermission('insights:reports:export', access)).toBe(false);
  });

  it('denies permissions when effective access is missing', () => {
    expect(canInsightsPermission('insights:reports:export', null)).toBe(false);
    expect(canInsightsPermission('insights:reports:create', undefined)).toBe(false);
  });
});
