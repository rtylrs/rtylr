import { describe, expect, it } from 'vitest';
import { normalizeFlowTeamMember } from './useFlowTeam';

describe('normalizeFlowTeamMember', () => {
  it('normalizes numeric ids into strings so picker selections resolve correctly', () => {
    const member = normalizeFlowTeamMember({
      id: 42,
      name: 'Alice',
      email: 'alice@example.com',
      role: 'member',
      apps: [],
      isActive: true,
    });

    expect(member.id).toBe('42');
  });

  it('prefers full name and avoids using email as a fake duplicate label source', () => {
    const member = normalizeFlowTeamMember({
      id: '7',
      name: '',
      fullName: 'Alice Smith',
      email: 'alice@example.com',
      role: 'member',
      apps: ['dash'],
      isActive: true,
    });

    expect(member.name).toBe('Alice Smith');
    expect(member.email).toBe('alice@example.com');
  });

  it('falls back to email only when no name fields exist', () => {
    const member = normalizeFlowTeamMember({
      id: '8',
      email: 'solo@example.com',
      role: 'member',
      apps: [],
      isActive: true,
    });

    expect(member.name).toBe('solo@example.com');
  });
});
