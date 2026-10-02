import { describe, expect, it } from 'vitest';
import type { EffectiveAccess } from '@voxire/types';
import {
  canDelete,
  canEditTask,
  canFlowPermission,
  hasFlowPermission,
  mostRestrictiveProjectRole,
} from './roleUtils';

function makeAccess(
  permissions: string[],
  overrides: Partial<Pick<EffectiveAccess, 'isOwner' | 'isAdmin' | 'subscribedApps'>> = {}
): EffectiveAccess {
  return {
    businessId: '12',
    dashPermissions: [],
    subscribedApps: overrides.subscribedApps ?? ['flow'],
    licenses: [],
    apps: [
      {
        appId: 'flow',
        licensed: true,
        hasAccess: true,
        permissions,
        permissionScopes: {},
        sources: ['explicit'],
      },
    ],
    isOwner: overrides.isOwner ?? false,
    isAdmin: overrides.isAdmin ?? false,
    authorizationVersion: 1,
  };
}

describe('Flow authorization', () => {
  it('uses the effective Flow permission set and denies when missing', () => {
    const access = makeAccess(['flow:projects:view']);

    expect(hasFlowPermission(access, 'flow:projects:view')).toBe(true);
    expect(canFlowPermission('flow:projects:create', access)).toBe(false);
    expect(canFlowPermission('flow:projects:view', null)).toBe(false);
  });

  it.each([
    ['owner', { isOwner: true }],
    ['admin', { isAdmin: true }],
  ] as const)('%s has full access inside subscribed apps', (_memberType, flags) => {
    const access = makeAccess([], flags);
    expect(canFlowPermission('flow:projects:edit', access)).toBe(true);
  });

  it('keeps owner/admin privileges within subscribed apps', () => {
    const access = makeAccess([], { isAdmin: true, subscribedApps: [] });
    expect(canFlowPermission('flow:projects:edit', access)).toBe(false);
  });
});

describe('project collaboration checks', () => {
  it('allows edits to project admins and members, but not viewers', () => {
    expect(canEditTask('admin')).toBe(true);
    expect(canEditTask('member')).toBe(true);
    expect(canEditTask('viewer')).toBe(false);
  });

  it('allows deletion to project admins only', () => {
    expect(canDelete('admin')).toBe(true);
    expect(canDelete('member')).toBe(false);
    expect(canDelete('viewer')).toBe(false);
  });

  it('returns the most restrictive role for multi-project operations', () => {
    expect(mostRestrictiveProjectRole([])).toBeUndefined();
    expect(mostRestrictiveProjectRole(['admin', 'viewer', 'member'])).toBe('viewer');
    expect(mostRestrictiveProjectRole(['member', 'admin'])).toBe('member');
  });
});
