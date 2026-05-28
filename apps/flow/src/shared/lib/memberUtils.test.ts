import { describe, it, expect } from 'vitest';

import {
  findMemberById,
  findMemberByEmail,
  memberDisplayName,
  memberRoleLabel,
  canEdit,
  isAdminOrOwner,
  sortMembersByRole,
  totalTeamCapacity,
} from './memberUtils';
import type { Member } from '../stores/flowStore';

function makeMember(overrides: Partial<Member>): Member {
  return {
    id: 'm1',
    name: 'Alice',
    email: 'alice@example.com',
    initials: 'AL',
    role: 'member',
    capacity: 40,
    ...overrides,
  } as Member;
}

describe('findMemberById', () => {
  it('finds member by id', () => {
    const members = [makeMember({ id: 'a' }), makeMember({ id: 'b', name: 'Bob' })];
    expect(findMemberById(members, 'b')?.name).toBe('Bob');
  });

  it('returns undefined for missing id', () => {
    expect(findMemberById([], 'a')).toBeUndefined();
  });
});

describe('findMemberByEmail', () => {
  it('finds member by email case-insensitively', () => {
    const members = [makeMember({ email: 'alice@example.com' })];
    expect(findMemberByEmail(members, 'ALICE@EXAMPLE.COM')?.name).toBe('Alice');
  });

  it('returns undefined when not found', () => {
    expect(findMemberByEmail([], 'x@y.com')).toBeUndefined();
  });
});

describe('memberDisplayName', () => {
  it('returns name when available', () => {
    expect(memberDisplayName(makeMember({ name: 'Alice' }))).toBe('Alice');
  });

  it('falls back to email when name is empty', () => {
    expect(memberDisplayName(makeMember({ name: '' }))).toBe('alice@example.com');
  });

  it('returns "Unknown" for null', () => {
    expect(memberDisplayName(null)).toBe('Unknown');
  });
});

describe('memberRoleLabel', () => {
  it('returns "Owner" for owner', () => {
    expect(memberRoleLabel('owner')).toBe('Owner');
  });

  it('returns "Viewer" for viewer', () => {
    expect(memberRoleLabel('viewer')).toBe('Viewer');
  });

  it('returns "Member" for member', () => {
    expect(memberRoleLabel('member')).toBe('Member');
  });
});

describe('canEdit', () => {
  it('returns true for owner', () => {
    expect(canEdit(makeMember({ role: 'owner' }))).toBe(true);
  });

  it('returns true for admin', () => {
    expect(canEdit(makeMember({ role: 'admin' }))).toBe(true);
  });

  it('returns true for member', () => {
    expect(canEdit(makeMember({ role: 'member' }))).toBe(true);
  });

  it('returns false for viewer', () => {
    expect(canEdit(makeMember({ role: 'viewer' }))).toBe(false);
  });
});

describe('isAdminOrOwner', () => {
  it('returns true for admin', () => {
    expect(isAdminOrOwner(makeMember({ role: 'admin' }))).toBe(true);
  });

  it('returns true for owner', () => {
    expect(isAdminOrOwner(makeMember({ role: 'owner' }))).toBe(true);
  });

  it('returns false for member', () => {
    expect(isAdminOrOwner(makeMember({ role: 'member' }))).toBe(false);
  });
});

describe('sortMembersByRole', () => {
  it('sorts owners first', () => {
    const members = [
      makeMember({ id: 'a', role: 'viewer' }),
      makeMember({ id: 'b', role: 'owner' }),
      makeMember({ id: 'c', role: 'admin' }),
      makeMember({ id: 'd', role: 'member' }),
    ];
    const sorted = sortMembersByRole(members);
    expect(sorted.map((m) => m.role)).toEqual(['owner', 'admin', 'member', 'viewer']);
  });

  it('does not mutate the original array', () => {
    const members = [makeMember({ role: 'viewer' }), makeMember({ role: 'owner' })];
    const original = [...members];
    sortMembersByRole(members);
    expect(members).toEqual(original);
  });
});

describe('totalTeamCapacity', () => {
  it('sums capacity across all members', () => {
    const members = [
      makeMember({ capacity: 40 }),
      makeMember({ capacity: 20 }),
      makeMember({ capacity: 30 }),
    ];
    expect(totalTeamCapacity(members)).toBe(90);
  });

  it('returns 0 for empty list', () => {
    expect(totalTeamCapacity([])).toBe(0);
  });
});
