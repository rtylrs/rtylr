import { describe, it, expect } from 'vitest';
import { canManageDash } from './canManageDash';
import type { StoredUser } from '../stores/userStore';
import { createEffectiveAccess } from '../../../../../test/effectiveAccess';

describe('canManageDash', () => {
  it('denies a missing user or an incomplete access snapshot', () => {
    expect(canManageDash(null)).toBe(false);
    expect(canManageDash(undefined)).toBe(false);
    expect(canManageDash({} as StoredUser)).toBe(false);
  });

  it.each([
    ['owner', createEffectiveAccess(undefined, { isOwner: true })],
    ['admin', createEffectiveAccess(undefined, { isAdmin: true })],
  ])('grants %s full Dash access', (_memberType, effectiveAccess) => {
    expect(canManageDash({ effectiveAccess } as StoredUser)).toBe(true);
  });

  it('honors explicit Dash team-management permissions', () => {
    const access = createEffectiveAccess(undefined, {
      dashPermissions: ['hub:team:edit'],
    });
    expect(canManageDash({ effectiveAccess: access } as StoredUser)).toBe(true);
  });

  it('does not infer Dash access from old role fields', () => {
    expect(canManageDash({ role: 'admin', dashRole: 'owner' } as StoredUser)).toBe(false);
  });
});
