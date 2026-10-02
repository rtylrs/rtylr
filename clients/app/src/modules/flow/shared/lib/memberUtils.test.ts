import { describe, it, expect } from 'vitest';

import {
  findMemberById,
  findMemberByEmail,
  memberDisplayName,
  totalTeamCapacity,
} from './memberUtils';
import type { Member } from '../stores/flowStore';

function makeMember(overrides: Partial<Member>): Member {
  return {
    id: 'm1',
    name: 'Alice',
    email: 'alice@example.com',
    initials: 'AL',
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
