import { describe, it, expect } from 'vitest';
import { canPurgeProject } from './canPurgeProject';

const project = { createdBy: '2001', ownerId: '3003' };

describe('canPurgeProject', () => {
  it('admits the creator', () => {
    expect(canPurgeProject(project, '2001', 'staff')).toBe(true);
  });

  it('admits the project owner', () => {
    expect(canPurgeProject(project, '3003', 'staff')).toBe(true);
  });

  it.each(['owner', 'admin', 'operator'] as const)('admits the %s bypass role', (role) => {
    expect(canPurgeProject(project, '9999', role)).toBe(true);
  });

  it.each(['billing', 'menu_manager', 'store_manager', 'staff', 'viewer'] as const)(
    'denies %s when they neither created nor own the project',
    (role) => {
      expect(canPurgeProject(project, '9999', role)).toBe(false);
    }
  );

  it('denies an unknown user id', () => {
    expect(canPurgeProject(project, null, 'staff')).toBe(false);
    expect(canPurgeProject(project, '', 'staff')).toBe(false);
  });

  it('denies a null role that is not the creator or owner', () => {
    expect(canPurgeProject(project, '9999', null)).toBe(false);
  });

  // A bypass role must not be defeated by a missing user id: the server admits
  // on role alone, so the client has to as well or an owner sees no button.
  it('admits a bypass role even without a resolved user id', () => {
    expect(canPurgeProject(project, null, 'owner')).toBe(true);
  });

  // An absent id must not match an absent user id into a false positive.
  it('denies a project whose creator and owner are unset', () => {
    expect(canPurgeProject({}, '2001', 'staff')).toBe(false);
  });
});
