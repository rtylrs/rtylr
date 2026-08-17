import type { Member } from '../stores/flowStore';
import type { FlowProjectRole } from '../services/flowApi';

/**
 * Two different role vocabularies live in this file and they are not
 * interchangeable:
 *
 * - `Member['role']` (owner|admin|member|viewer) is the business-wide team role.
 *   Workspace-level questions - managing the workspace, creating projects,
 *   inviting people - are answered with it.
 * - `FlowProjectRole` (admin|member|viewer) comes from `flow_project_members` and
 *   answers per-project questions. Editing and deleting inside a project are
 *   scoped to a single project, so they use this one.
 *
 * A company admin who is only a viewer on a project must not be able to edit its
 * tasks, which is why the two must stay separate.
 */

/**
 * Returns the i18n key for a member role's display label. Role names are
 * user-facing copy, not business logic, so this hands back a translation key
 * rather than an English string - callers render it with `t(roleLabel(role))`.
 */
export function roleLabel(role: Member['role']): string {
  const keys: Record<Member['role'], string> = {
    owner: 'common.roles.owner',
    admin: 'common.roles.admin',
    member: 'common.roles.member',
    viewer: 'common.roles.viewer',
  };
  return keys[role];
}

/** Returns true if the role can manage workspace settings. */
export function canManageWorkspace(role: Member['role']): boolean {
  return role === 'owner' || role === 'admin';
}

/** Returns true if the role can create projects. */
export function canCreateProject(role: Member['role']): boolean {
  return role === 'owner' || role === 'admin' || role === 'member';
}

/** Returns true if the project role can edit tasks in that project. */
export function canEditTask(role: FlowProjectRole): boolean {
  return role !== 'viewer';
}

/** Returns true if the project role can delete resources in that project. */
export function canDelete(role: FlowProjectRole): boolean {
  return role === 'admin';
}

/** Project roles ordered most to least privileged. */
const PROJECT_ROLE_RANK: Record<FlowProjectRole, number> = {
  admin: 0,
  member: 1,
  viewer: 2,
};

/**
 * Returns the least privileged of the given project roles, or undefined for an
 * empty list. Used to gate an action that spans several projects at once: the
 * user may only do it if they may do it everywhere it would apply.
 */
export function mostRestrictiveProjectRole(roles: FlowProjectRole[]): FlowProjectRole | undefined {
  return roles.reduce<FlowProjectRole | undefined>(
    (lowest, role) =>
      lowest === undefined || PROJECT_ROLE_RANK[role] > PROJECT_ROLE_RANK[lowest] ? role : lowest,
    undefined
  );
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
