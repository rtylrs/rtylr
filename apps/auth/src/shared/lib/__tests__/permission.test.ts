import { describe, it, expect } from 'vitest';
import {
  hasRole,
  hasPermissions,
  hasAnyPermission,
  canAccess,
  mergePermissions,
} from '../permission';

describe('hasRole', () => {
  it('returns true for an allowed role', () => {
    expect(hasRole('admin', ['admin', 'editor'])).toBe(true);
  });

  it('returns false for a disallowed role', () => {
    expect(hasRole('viewer', ['admin', 'editor'])).toBe(false);
  });
});

describe('hasPermissions', () => {
  it('returns true when all required permissions are present', () => {
    expect(hasPermissions(['read', 'write', 'delete'], ['read', 'write'])).toBe(true);
  });

  it('returns false when a required permission is missing', () => {
    expect(hasPermissions(['read'], ['read', 'write'])).toBe(false);
  });

  it('returns true for empty required list', () => {
    expect(hasPermissions(['read'], [])).toBe(true);
  });
});

describe('hasAnyPermission', () => {
  it('returns true if at least one permission matches', () => {
    expect(hasAnyPermission(['read', 'write'], ['write', 'delete'])).toBe(true);
  });

  it('returns false if no permissions match', () => {
    expect(hasAnyPermission(['read'], ['write', 'delete'])).toBe(false);
  });
});

describe('canAccess', () => {
  it('returns true when role and permissions both match', () => {
    expect(
      canAccess('admin', ['read', 'write'], {
        roles: ['admin'],
        permissions: ['read'],
      })
    ).toBe(true);
  });

  it('returns false when role does not match', () => {
    expect(
      canAccess('viewer', ['read', 'write'], {
        roles: ['admin'],
        permissions: ['read'],
      })
    ).toBe(false);
  });

  it('returns false when permissions are insufficient', () => {
    expect(
      canAccess('admin', ['read'], {
        roles: ['admin'],
        permissions: ['read', 'write'],
      })
    ).toBe(false);
  });
});

describe('mergePermissions', () => {
  it('combines two lists without duplicates', () => {
    const merged = mergePermissions(['read', 'write'], ['write', 'delete']);
    expect(merged.sort()).toEqual(['delete', 'read', 'write']);
  });

  it('handles empty arrays', () => {
    expect(mergePermissions([], ['read'])).toEqual(['read']);
  });
});
