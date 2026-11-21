import { describe, it, expect } from 'vitest';

import {
  assigneeOptions,
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

describe('assigneeOptions', () => {
  const members = [makeMember({}), makeMember({ id: 'm2', name: 'Bob', initials: 'BO' })];

  it('offers nobody first, then the members', () => {
    expect(assigneeOptions(members, { id: '', name: '' }, 'Unassigned')).toEqual([
      { id: '', name: 'Unassigned', initials: '' },
      { id: 'm1', name: 'Alice', initials: 'AL' },
      { id: 'm2', name: 'Bob', initials: 'BO' },
    ]);
  });

  it('keeps someone assigned who has left the member list', () => {
    const options = assigneeOptions(members, { id: 'm9', name: 'Cara', initials: 'CA' }, 'None');
    expect(options.map((o) => o.id)).toEqual(['', 'm9', 'm1', 'm2']);
    expect(options[1]).toEqual({ id: 'm9', name: 'Cara', initials: 'CA' });
  });

  it('does not repeat an assignee who is still a member', () => {
    const options = assigneeOptions(members, { id: 'm2', name: 'Bob' }, 'None');
    expect(options.map((o) => o.id)).toEqual(['', 'm1', 'm2']);
  });
});
