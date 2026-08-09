import { describe, it, expect } from 'vitest';

import { findProjectRole } from './currentUser';
import type { ApiProjectMember } from '../services/flowApi';

function makeProjectMember(overrides: Partial<ApiProjectMember>): ApiProjectMember {
  return {
    id: 'pm1',
    projectId: 'p1',
    userId: '7',
    name: 'Alice',
    email: 'alice@example.com',
    role: 'member',
    addedAt: '2026-01-01T00:00:00Z',
    ...overrides,
  };
}

describe('findProjectRole', () => {
  it('returns the signed-in user role on the project', () => {
    const members = [
      makeProjectMember({ userId: '7', role: 'viewer' }),
      makeProjectMember({ id: 'pm2', userId: '8', role: 'admin' }),
    ];
    expect(findProjectRole(members, '7')).toBe('viewer');
    expect(findProjectRole(members, '8')).toBe('admin');
  });

  it('fails closed while the member list is still loading', () => {
    expect(findProjectRole(undefined, '7')).toBeUndefined();
  });

  it('fails closed when there is no signed-in user', () => {
    expect(findProjectRole([makeProjectMember({})], null)).toBeUndefined();
  });

  it('fails closed when the user is not a member of the project', () => {
    expect(findProjectRole([makeProjectMember({ userId: '7' })], '99')).toBeUndefined();
  });

  it('matches ids across string and number representations', () => {
    // The API layer can hand back a numeric id even though the contract types it
    // as a string, so a project role must not be missed on a type mismatch.
    const members = [{ ...makeProjectMember({ role: 'admin' }), userId: 7 as unknown as string }];
    expect(findProjectRole(members, '7')).toBe('admin');
  });
});
