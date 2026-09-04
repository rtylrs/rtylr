import { describe, it, expect } from 'vitest';
import { isAdminOrOwner } from './requireAdminOrOwner';
import type { StoredUser } from '../stores/userStore';
import type { EffectiveAccess } from '@voxire/types';

const baseAccess: EffectiveAccess = {
  businessId: '1',
  dashRole: 'member',
  dashRoleId: null,
  dashPermissions: [],
  subscribedApps: [],
  licenses: [],
  apps: [],
  isOwner: false,
  isAdmin: false,
  authorizationVersion: 1,
};

describe('isAdminOrOwner', () => {
  it('denies null and undefined users', () => {
    expect(isAdminOrOwner(null)).toBe(false);
    expect(isAdminOrOwner(undefined)).toBe(false);
  });

  it('passes explicit owners', () => {
    expect(isAdminOrOwner({ isOwner: true, role: 'staff' } as StoredUser)).toBe(true);
  });

  it('denies when isOwner is omitted and no access document is present', () => {
    expect(isAdminOrOwner({ role: 'staff' } as StoredUser)).toBe(false);
    expect(isAdminOrOwner({} as StoredUser)).toBe(false);
  });

  it('denies explicit non-owners without a grant', () => {
    expect(isAdminOrOwner({ isOwner: false, role: 'staff' } as StoredUser)).toBe(false);
  });

  it('honors management grants from the access document', () => {
    for (const permission of ['hub:team:edit', 'hub:roles:view', 'hub:licenses:manage']) {
      expect(
        isAdminOrOwner({
          isOwner: false,
          role: 'staff',
          effectiveAccess: { ...baseAccess, dashPermissions: [permission] },
        } as StoredUser)
      ).toBe(true);
    }
    expect(
      isAdminOrOwner({
        isOwner: false,
        role: 'staff',
        effectiveAccess: { ...baseAccess, dashPermissions: ['hub:orders:view'] },
      } as StoredUser)
    ).toBe(false);
  });

  it('falls back to legacy role and dashRole without an access document', () => {
    expect(isAdminOrOwner({ isOwner: false, role: 'admin' } as StoredUser)).toBe(true);
    expect(isAdminOrOwner({ isOwner: false, role: 'Owner' } as StoredUser)).toBe(true);
    expect(isAdminOrOwner({ isOwner: false, dashRole: 'admin' } as StoredUser)).toBe(true);
    expect(isAdminOrOwner({ isOwner: false, role: 'staff' } as StoredUser)).toBe(false);
  });
});
