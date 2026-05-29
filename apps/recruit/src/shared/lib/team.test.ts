import { describe, it, expect } from 'vitest';
import {
  recruiterRoleLabel,
  totalTeamHires,
  avgPositionsPerRecruiter,
  topRecruiterByHires,
  membersByRole,
  allRecruiterRoles,
} from './team';
import type { RecruiterMember } from './team';

function makeMember(
  id: string,
  role: 'recruiter' | 'coordinator' | 'sourcer' | 'manager' | 'interviewer',
  activePositions: number,
  hiredThisMonth: number
): RecruiterMember {
  return { id, name: `Member ${id}`, role, activePositions, hiredThisMonth };
}

const members = [
  makeMember('1', 'recruiter', 5, 3),
  makeMember('2', 'recruiter', 8, 5),
  makeMember('3', 'coordinator', 0, 0),
  makeMember('4', 'sourcer', 3, 1),
];

describe('recruiterRoleLabel', () => {
  it('returns display label for known roles', () => {
    expect(recruiterRoleLabel('recruiter')).toBe('Recruiter');
    expect(recruiterRoleLabel('coordinator')).toBe('Recruiting Coordinator');
  });

  it('returns raw value for unknown role', () => {
    expect(recruiterRoleLabel('contractor')).toBe('contractor');
  });
});

describe('totalTeamHires', () => {
  it('sums hires across all members', () => {
    expect(totalTeamHires(members)).toBe(9);
  });
});

describe('avgPositionsPerRecruiter', () => {
  it('returns average active positions', () => {
    expect(avgPositionsPerRecruiter(members)).toBe(4);
  });

  it('returns 0 for empty team', () => {
    expect(avgPositionsPerRecruiter([])).toBe(0);
  });
});

describe('topRecruiterByHires', () => {
  it('returns member with most hires', () => {
    expect(topRecruiterByHires(members)?.id).toBe('2');
  });

  it('returns undefined for empty array', () => {
    expect(topRecruiterByHires([])).toBeUndefined();
  });
});

describe('membersByRole', () => {
  it('groups members by role', () => {
    const grouped = membersByRole(members);
    expect(grouped['recruiter']).toHaveLength(2);
    expect(grouped['sourcer']).toHaveLength(1);
  });
});

describe('allRecruiterRoles', () => {
  it('returns all defined roles', () => {
    const roles = allRecruiterRoles();
    expect(roles).toContain('recruiter');
    expect(roles).toContain('manager');
    expect(roles).toHaveLength(5);
  });
});
