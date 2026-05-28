/**
 * Team member utility helpers for the Dash app.
 */

export type TeamRole = 'owner' | 'admin' | 'manager' | 'staff' | 'viewer';

export interface TeamMember {
  id: string;
  name: string;
  email: string;
  role: TeamRole;
  active?: boolean;
}

/** Return the display label for a role. */
export function roleLabel(role: TeamRole): string {
  const labels: Record<TeamRole, string> = {
    owner: 'Owner',
    admin: 'Admin',
    manager: 'Manager',
    staff: 'Staff',
    viewer: 'Viewer',
  };
  return labels[role] ?? role;
}

/** Return true if the role has admin-level access. */
export function isAdminRole(role: TeamRole): boolean {
  return role === 'owner' || role === 'admin';
}

/** Return true if the role can manage other users. */
export function canManageUsers(role: TeamRole): boolean {
  return role === 'owner' || role === 'admin' || role === 'manager';
}

/** Sort team members: owners first, then by role weight, then alphabetically. */
export function sortTeamMembers(members: TeamMember[]): TeamMember[] {
  const roleOrder: Record<TeamRole, number> = {
    owner: 0,
    admin: 1,
    manager: 2,
    staff: 3,
    viewer: 4,
  };
  return [...members].sort((a, b) => {
    const diff = roleOrder[a.role] - roleOrder[b.role];
    if (diff !== 0) return diff;
    return a.name.localeCompare(b.name);
  });
}

/** Filter members to only active ones. */
export function activeMembers(members: TeamMember[]): TeamMember[] {
  return members.filter((m) => m.active !== false);
}

/** Find a member by email (case-insensitive). */
export function findByEmail(members: TeamMember[], email: string): TeamMember | undefined {
  const lower = email.toLowerCase();
  return members.find((m) => m.email.toLowerCase() === lower);
}

/** Count members grouped by role. */
export function countByRole(members: TeamMember[]): Record<TeamRole, number> {
  const counts: Record<TeamRole, number> = {
    owner: 0,
    admin: 0,
    manager: 0,
    staff: 0,
    viewer: 0,
  };
  for (const m of members) {
    if (m.role in counts) counts[m.role]++;
  }
  return counts;
}
