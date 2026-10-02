import { describe, expect, it } from 'vitest';
import { normalizeFlowTeamMember } from './useFlowTeam';

describe('normalizeFlowTeamMember', () => {
  it('prefers full name and avoids using email as a fake duplicate label source', () => {
    const member = normalizeFlowTeamMember({
      id: 7,
      fullName: 'Alice Smith',
      email: 'alice@example.com',
      memberType: 'member',
      licensedAppIds: ['flow'],
      permissionProfileAssignments: [],
      effectiveAccess: {
        businessId: '1',
        dashPermissions: [],
        subscribedApps: ['flow'],
        licenses: [],
        apps: [
          {
            appId: 'flow',
            licensed: true,
            hasAccess: true,
            permissions: ['flow:projects:view'],
            permissionScopes: [{ permissionKey: 'flow:projects:view', scopes: [] }],
            sources: ['explicit'],
          },
        ],
        isOwner: false,
        isAdmin: false,
        authorizationVersion: '1',
      },
      isActive: true,
      createdAt: '2026-01-01',
    });

    expect(member.id).toBe('7');
    expect(member.name).toBe('Alice Smith');
    expect(member.email).toBe('alice@example.com');
    expect(member.hasFlowAccess).toBe(true);
  });

  it('falls back to email only when no full name exists', () => {
    const member = normalizeFlowTeamMember({
      id: 8,
      fullName: '',
      email: 'solo@example.com',
      memberType: 'member',
      licensedAppIds: [],
      permissionProfileAssignments: [],
      effectiveAccess: null,
      isActive: true,
      createdAt: '2026-01-01',
    });

    expect(member.name).toBe('solo@example.com');
    expect(member.hasFlowAccess).toBe(false);
  });
});
