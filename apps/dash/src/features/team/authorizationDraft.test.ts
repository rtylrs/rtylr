import { describe, it, expect } from 'vitest';
import { legacyRoleToDashRole, dashRoleToLegacyRole } from './authorizationDraft';

describe('legacyRoleToDashRole', () => {
  it('preserves owners instead of demoting them to members', () => {
    expect(legacyRoleToDashRole('owner')).toBe('owner');
  });

  it('matches case-insensitively', () => {
    expect(legacyRoleToDashRole('Admin')).toBe('admin');
    expect(legacyRoleToDashRole('OWNER')).toBe('owner');
    expect(legacyRoleToDashRole('Viewer')).toBe('viewer');
  });

  it('maps known legacy roles', () => {
    expect(legacyRoleToDashRole('admin')).toBe('admin');
    expect(legacyRoleToDashRole('operator')).toBe('manager');
    expect(legacyRoleToDashRole('billing')).toBe('manager');
    expect(legacyRoleToDashRole('menu_manager')).toBe('manager');
    expect(legacyRoleToDashRole('store_manager')).toBe('manager');
    expect(legacyRoleToDashRole('viewer')).toBe('viewer');
    expect(legacyRoleToDashRole('staff')).toBe('member');
    expect(legacyRoleToDashRole('something-else')).toBe('member');
  });
});

describe('dashRoleToLegacyRole', () => {
  it('round-trips the editable roles without touching owner display', () => {
    expect(dashRoleToLegacyRole('admin')).toBe('admin');
    expect(dashRoleToLegacyRole('manager')).toBe('operator');
    expect(dashRoleToLegacyRole('viewer')).toBe('viewer');
    expect(dashRoleToLegacyRole('member')).toBe('staff');
  });
});
