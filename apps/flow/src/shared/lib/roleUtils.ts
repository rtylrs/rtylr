import type { Member } from '../stores/flowStore';

/** Returns a human-readable label for a member role. */
export function roleLabel(role: Member['role']): string {
  const labels: Record<Member['role'], string> = {
    owner: 'Owner',
    admin: 'Admin',
    member: 'Member',
    viewer: 'Viewer',
  };
  return labels[role];
}

/** Returns true if the role can manage workspace settings. */
export function canManageWorkspace(role: Member['role']): boolean {
  return role === 'owner' || role === 'admin';
}

/** Returns true if the role can create projects. */
export function canCreateProject(role: Member['role']): boolean {
  return role === 'owner' || role === 'admin' || role === 'member';
}

/** Returns true if the role can edit tasks. */
export function canEditTask(role: Member['role']): boolean {
  return role !== 'viewer';
}

/** Returns true if the role can delete resources. */
export function canDelete(role: Member['role']): boolean {
  return role === 'owner' || role === 'admin';
}

/** Returns true if the role can invite members. */
export function canInviteMembers(role: Member['role']): boolean {
  return role === 'owner' || role === 'admin';
}

/** Returns the numeric priority of a role (owner highest). */
export function rolePriority(role: Member['role']): number {
  const priorities: Record<Member['role'], number> = {
    owner: 0,
    admin: 1,
    member: 2,
    viewer: 3,
  };
  return priorities[role];
}

/** Sorts members by role priority (owners first). */
export function sortMembersByRole(members: Member[]): Member[] {
  return [...members].sort((a, b) => rolePriority(a.role) - rolePriority(b.role));
}

/** Returns members with a specific role. */
export function membersByRole(members: Member[], role: Member['role']): Member[] {
  return members.filter((m) => m.role === role);
}
