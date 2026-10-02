import { useQueries } from '@tanstack/react-query';
import { useUserStore } from '../stores/userStore';
import { canEditTask, canDelete, mostRestrictiveProjectRole, canFlowPermission } from './roleUtils';
import {
  projectMembersQueryOptions,
  useProjectMembers,
} from '../../features/projects/useProjectMembers';
import { useProject } from '../../features/projects-v2/useAllProjects';
import type { ApiProjectMember, FlowProjectRole } from '../services/flowApi';

export function getCurrentUserId(): string | null {
  const id = useUserStore.getState().user?.id;
  return id === undefined ? null : String(id);
}

/**
 * Whether the task counts as "mine" for My tasks.
 *
 * Compared as strings, for the reason findProjectRole below already records:
 * the API layer can hand back a numeric id even where the contract types it
 * as a string. A strict === between a number and a string is false for every
 * task, and this predicate gates the My tasks count, the Today count and the
 * whole Today list, so it fails silently and completely.
 */
export function isTaskMine(taskAssigneeId: string): boolean {
  const id = getCurrentUserId();
  return id !== null && String(taskAssigneeId) === String(id);
}

/** The signed-in user's id, reactive to store updates. */
function useCurrentUserId(): string | null {
  return useUserStore((state) => (state.user?.id === undefined ? null : String(state.user.id)));
}

/**
 * Picks the signed-in user's role out of a project's member list.
 * Ids are compared as strings because the API layer can hand back a numeric id
 * even where the contract types it as a string.
 * Returns undefined when the list hasn't loaded or the user isn't a member.
 */
export function findProjectRole(
  members: ApiProjectMember[] | undefined,
  userId: string | null
): FlowProjectRole | undefined {
  if (!members || userId === null) return undefined;
  return members.find((m: { userId: string }) => String(m.userId) === String(userId))?.role;
}

/** Owners and admins have full access; members also need a project role. */
export function useCurrentProjectRole(projectId?: string): FlowProjectRole | undefined {
  const userId = useCurrentUserId();
  const access = useUserStore((state) => state.user?.effectiveAccess);
  const readableProjectId = canFlowPermission('flow:projects:view', access) ? projectId : undefined;
  const { data: members } = useProjectMembers(readableProjectId);
  const { data: project } = useProject(readableProjectId);
  if (access?.isOwner || access?.isAdmin) return 'admin';
  const memberRole = findProjectRole(members, userId);
  if (memberRole !== undefined) return memberRole;
  if (
    userId !== null &&
    project !== undefined &&
    (String(project.ownerId) === String(userId) || String(project.createdBy) === String(userId))
  ) {
    return 'admin';
  }
  return undefined;
}

/**
 * The least privileged role the signed-in user holds across several projects.
 *
 * Task selection in Flow is global (`flowStore.selectedTaskIds`) and cross-project
 * views like Triage and My issues let a selection span projects, so a bulk action
 * can touch more than one project at once. Gating on the most restrictive role
 * means the user may only bulk-act when they'd be allowed to act in every project
 * the action would reach.
 *
 * Returns undefined for an empty list, or when any one project's role is still
 * loading or the user isn't a member of it (fail closed).
 */
export function useMostRestrictiveProjectRole(projectIds: string[]): FlowProjectRole | undefined {
  const userId = useCurrentUserId();
  const access = useUserStore((state) => state.user?.effectiveAccess);
  const results = useQueries({
    queries: projectIds.map((id) => ({
      ...projectMembersQueryOptions(id),
      enabled: canFlowPermission('flow:projects:view', access),
    })),
  });

  if (projectIds.length === 0) return undefined;
  if (access?.isOwner || access?.isAdmin) return 'admin';

  const roles: FlowProjectRole[] = [];
  for (const result of results) {
    const role = findProjectRole(result.data, userId);
    if (role === undefined) return undefined;
    roles.push(role);
  }
  return mostRestrictiveProjectRole(roles);
}

/** Whether the current user can edit task fields (title, status, assignee, etc) in a project. */
export function useCanEditTask(projectId?: string): boolean {
  const projectRole = useCurrentProjectRole(projectId);
  const access = useUserStore((state) => state.user?.effectiveAccess);
  return (
    canFlowPermission('flow:tasks:edit', access) &&
    projectRole !== undefined &&
    canEditTask(projectRole)
  );
}

/**
 * Whether the current user may perform project-admin operations on a project.
 *
 * This is onion's `RequireProjectRole(..., FlowRoleAdmin)`: the gate behind
 * deleting, archiving, restoring and flipping visibility alike. Named for the
 * authority rather than one of its uses so the next admin-gated control does
 * not grow a second copy of the same check.
 */
export function useCanAdministerProject(projectId?: string): boolean {
  const role = useCurrentProjectRole(projectId);
  const access = useUserStore((state) => state.user?.effectiveAccess);
  return canFlowPermission('flow:projects:edit', access) && role !== undefined && canDelete(role);
}

/** Whether the current user can delete resources in a project (project admins only). */
export function useCanDeleteResource(projectId?: string): boolean {
  return useCanAdministerProject(projectId);
}
