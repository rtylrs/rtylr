import type { Member } from '../stores/flowStore';

/** Derive up-to-two-letter initials from a display name (or email). */
export function initialsFromName(name: string): string {
  const parts = name.trim().split(/\s+/).filter(Boolean);
  if (parts.length === 0) return '?';
  if (parts.length === 1) return parts[0].slice(0, 2).toUpperCase();
  return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
}

/** Find a member by their ID. */
export function findMemberById(members: Member[], id: string): Member | undefined {
  return members.find((m) => m.id === id);
}

/** Find a member by their email (case-insensitive). */
export function findMemberByEmail(members: Member[], email: string): Member | undefined {
  const lower = email.toLowerCase();
  return members.find((m) => m.email.toLowerCase() === lower);
}

/** Returns the display name for a member (name or email fallback). */
export function memberDisplayName(member: Member | null | undefined): string {
  if (!member) return 'Unknown';
  return member.name?.trim() || member.email;
}

/** Returns the role label for a member role. */
export function memberRoleLabel(role: Member['role']): string {
  const labels: Record<Member['role'], string> = {
    owner: 'Owner',
    admin: 'Admin',
    member: 'Member',
    viewer: 'Viewer',
  };
  return labels[role] ?? role;
}

/** Returns true when the member can edit (is owner, admin, or member). */
export function canEdit(member: Member): boolean {
  return member.role === 'owner' || member.role === 'admin' || member.role === 'member';
}

/** Returns true when the member is an admin or owner. */
export function isAdminOrOwner(member: Member): boolean {
  return member.role === 'owner' || member.role === 'admin';
}

/** Sort members: owners first, then admins, then members, then viewers. */
export function sortMembersByRole(members: Member[]): Member[] {
  const order: Record<Member['role'], number> = { owner: 0, admin: 1, member: 2, viewer: 3 };
  return [...members].sort((a, b) => order[a.role] - order[b.role]);
}

/** Total available capacity (hours/week) across all members. */
export function totalTeamCapacity(members: Member[]): number {
  return members.reduce((sum, m) => sum + (m.capacity ?? 0), 0);
}
