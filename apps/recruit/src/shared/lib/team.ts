// Recruiting team helpers for the recruit admin panel.

export type RecruiterRole = 'recruiter' | 'coordinator' | 'sourcer' | 'manager' | 'interviewer';

export interface RecruiterMember {
  id: string;
  name: string;
  role: RecruiterRole;
  activePositions: number;
  hiredThisMonth: number;
}

/**
 * recruiterRoleLabel returns a display-friendly label for a recruiter role.
 */
export function recruiterRoleLabel(role: RecruiterRole | string): string {
  const labels: Record<string, string> = {
    recruiter: 'Recruiter',
    coordinator: 'Recruiting Coordinator',
    sourcer: 'Sourcer',
    manager: 'Recruiting Manager',
    interviewer: 'Interviewer',
  };
  return labels[role] ?? role;
}

/**
 * totalTeamHires returns the sum of hires across all team members this month.
 */
export function totalTeamHires(members: RecruiterMember[]): number {
  return members.reduce((sum, m) => sum + m.hiredThisMonth, 0);
}

/**
 * avgPositionsPerRecruiter returns the average active positions per team member.
 */
export function avgPositionsPerRecruiter(members: RecruiterMember[]): number {
  if (members.length === 0) return 0;
  const total = members.reduce((sum, m) => sum + m.activePositions, 0);
  return Math.round((total / members.length) * 10) / 10;
}

/**
 * topRecruiterByHires returns the team member with the most hires this month.
 */
export function topRecruiterByHires(members: RecruiterMember[]): RecruiterMember | undefined {
  if (members.length === 0) return undefined;
  return members.reduce((best, m) => (m.hiredThisMonth > best.hiredThisMonth ? m : best));
}

/**
 * membersByRole groups team members by their role.
 */
export function membersByRole(members: RecruiterMember[]): Record<string, RecruiterMember[]> {
  return members.reduce<Record<string, RecruiterMember[]>>((acc, m) => {
    acc[m.role] = [...(acc[m.role] ?? []), m];
    return acc;
  }, {});
}

/**
 * allRecruiterRoles returns all defined recruiter role values.
 */
export function allRecruiterRoles(): RecruiterRole[] {
  return ['recruiter', 'coordinator', 'sourcer', 'manager', 'interviewer'];
}
