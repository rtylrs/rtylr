import { describe, expect, it } from 'vitest';
import { normalizeFlowTeamMember } from './useFlowTeam';

describe('normalizeFlowTeamMember', () => {
  it('prefers full name and avoids using email as a fake duplicate label source', () => {
    const member = normalizeFlowTeamMember({
      id: '7',
      fullName: 'Alice Smith',
      email: 'alice@example.com',
      role: 'member',
      apps: ['dash'],
      isActive: true,
      createdAt: '2026-01-01',
    });

    expect(member.id).toBe('7');
    expect(member.name).toBe('Alice Smith');
    expect(member.email).toBe('alice@example.com');
  });

  it('falls back to email only when no full name exists', () => {
    const member = normalizeFlowTeamMember({
      id: '8',
      fullName: '',
      email: 'solo@example.com',
      role: 'member',
      apps: [],
      isActive: true,
      createdAt: '2026-01-01',
    });

    expect(member.name).toBe('solo@example.com');
  });
});
