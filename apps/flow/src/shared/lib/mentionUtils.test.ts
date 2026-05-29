import { describe, it, expect } from 'vitest';

import {
  extractMentions,
  hasMentions,
  resolveMentions,
  membersForMentions,
  buildMention,
  mentionCandidates,
} from './mentionUtils';
import type { Member } from '../stores/flowStore';

function makeMember(overrides: Partial<Member>): Member {
  return {
    id: 'm1',
    name: 'Alice Smith',
    email: 'alice@example.com',
    initials: 'AS',
    role: 'member',
    capacity: 40,
    ...overrides,
  } as Member;
}

describe('extractMentions', () => {
  it('extracts @handles from text', () => {
    expect(extractMentions('Hey @alice.smith, check this')).toEqual(['@alice.smith']);
  });

  it('returns empty array for no mentions', () => {
    expect(extractMentions('No mentions here')).toEqual([]);
  });

  it('deduplicates mentions', () => {
    expect(extractMentions('@alice @alice @bob')).toEqual(['@alice', '@bob']);
  });
});

describe('hasMentions', () => {
  it('returns true when text has mentions', () => {
    expect(hasMentions('Hi @alice')).toBe(true);
  });

  it('returns false with no mentions', () => {
    expect(hasMentions('No mentions')).toBe(false);
  });
});

describe('resolveMentions', () => {
  it('replaces @handle with member display name', () => {
    const members = [makeMember({ name: 'Alice Smith' })];
    const result = resolveMentions('Hey @alice.smith!', members);
    expect(result).toContain('@Alice Smith');
  });

  it('leaves unresolved mentions unchanged', () => {
    const result = resolveMentions('Hey @unknown', []);
    expect(result).toBe('Hey @unknown');
  });
});

describe('membersForMentions', () => {
  it('returns members matching the mentions', () => {
    const members = [
      makeMember({ name: 'Alice Smith' }),
      makeMember({ id: 'm2', name: 'Bob Jones' }),
    ];
    const result = membersForMentions('Hello @alice.smith', members);
    expect(result.length).toBe(1);
    expect(result[0].name).toBe('Alice Smith');
  });
});

describe('buildMention', () => {
  it('creates @first.last mention', () => {
    const member = makeMember({ name: 'Alice Smith' });
    expect(buildMention(member)).toBe('@alice.smith');
  });

  it('lowercases the mention', () => {
    const member = makeMember({ name: 'BOB JONES' });
    expect(buildMention(member)).toBe('@bob.jones');
  });
});

describe('mentionCandidates', () => {
  it('filters by query', () => {
    const members = [
      makeMember({ name: 'Alice Smith' }),
      makeMember({ id: 'm2', name: 'Bob Jones', email: 'bob@example.com' }),
    ];
    expect(mentionCandidates(members, 'alice').length).toBe(1);
  });

  it('returns all members for empty query', () => {
    const members = [makeMember({}), makeMember({ id: 'm2', name: 'Bob' })];
    expect(mentionCandidates(members, '').length).toBe(2);
  });

  it('sorts alphabetically by name', () => {
    const members = [
      makeMember({ id: 'm2', name: 'Zara' }),
      makeMember({ id: 'm1', name: 'Alice Smith' }),
    ];
    expect(mentionCandidates(members, '')[0].name).toBe('Alice Smith');
  });
});
