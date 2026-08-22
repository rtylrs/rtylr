import { useQueries } from '@tanstack/react-query';
import { useUserStore as useAuthUserStore } from '@voxire/auth';
import { useUserStore } from '../stores/userStore';
import { useFlowStore } from '../stores/flowStore';
import {
  canEditTask,
  canDelete,
  hasBusinessRoleBypass,
  mostRestrictiveProjectRole,
} from './roleUtils';
import {
  projectMembersQueryOptions,
  useProjectMembers,
} from '../../features/projects/useProjectMembers';
import type { ApiProjectMember, FlowProjectRole } from '../services/flowApi';

/** Fallback user id when no authenticated user is available (dev only) */
export const DEMO_USER_ID = 'u1';

export function getCurrentUserId(): string | null {
  const userId = useUserStore.getState().user?.id;
  if (userId !== undefined) return userId;
  if (import.meta.env.PROD) return null;
  return DEMO_USER_ID;
}

/** Whether the task counts as "mine" for My tasks. */
export function isTaskMine(taskAssigneeId: string): boolean {
  const id = getCurrentUserId();
  return id !== null && taskAssigneeId === id;
}

/** The signed-in user's id, reactive to store updates. */
function useCurrentUserId(): string | null {
  return useUserStore((s) => s.user?.id) ?? (import.meta.env.PROD ? null : DEMO_USER_ID);
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

/**
 * The signed-in user's role on one project, from `flow_project_members`.
 *
 * This is deliberately not the business-wide team role: someone can be a company
 * admin and still only a viewer on a given project, and the project role is the
 * one that governs what they may do inside it.
 *
 * Returns undefined while the member list is loading, when no projectId is known
 * yet, or when the user isn't a member of the project - callers must treat that
 * as "not permitted" (fail closed).
 */
export function useCurrentProjectRole(projectId?: string): FlowProjectRole | undefined {
  const userId = useCurrentUserId();
  const { data: members } = useProjectMembers(projectId);
  return findProjectRole(members, userId);
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
  const results = useQueries({
    queries: projectIds.map((id) => projectMembersQueryOptions(id)),
  });

  if (projectIds.length === 0) return undefined;

  const roles: FlowProjectRole[] = [];
  for (const result of results) {
    const role = findProjectRole(result.data, userId);
    if (role === undefined) return undefined;
    roles.push(role);
  }
  return mostRestrictiveProjectRole(roles);
}

/**
 * The signed-in user's role in the active workspace's member list
 * (`flow_workspace_members`), scoped to `useCanEditTask` below as a
 * fallback only - never a standalone permission source.
 *
 * onion's own task-update authorization (`RequireProjectRole` /
 * `projectRoleFor`) already admits any member of a project's workspace,
 * not just explicit `flow_project_members` rows. Without this fallback,
 * the frontend was stricter than the backend it talks to: a workspace
 * member who was never individually added to one of that workspace's
 * projects saw every edit control disabled here, even though the same
 * write would have succeeded had the button not been disabled.
 */
function useCurrentWorkspaceRole(): string | undefined {
  // Workspace layer removed; project roles are the only permission source.
  return undefined;
}

/**
 * A business-wide owner/admin/operator should never be blocked from acting on
 * their own business's data by a per-project or per-workspace membership gap -
 * onion's authorization is gaining the same owner/admin/operator bypass, so
 * withholding it here would only recreate the disabled-but-would-succeed
 * mismatch this codebase otherwise guards against, just in the other direction.
 */
export function useHasBusinessRoleBypass(): boolean {
  return hasBusinessRoleBypass(useAuthUserStore((s) => s.role));
}

/** Whether the current user can edit task fields (title, status, assignee, etc) in a project. */
export function useCanEditTask(projectId?: string): boolean {
  const hasBusinessRoleBypass = useHasBusinessRoleBypass();
  const projectRole = useCurrentProjectRole(projectId);
  const workspaceRole = useCurrentWorkspaceRole();
  if (hasBusinessRoleBypass) return true;
  if (projectRole !== undefined) return canEditTask(projectRole);
  return workspaceRole !== undefined && workspaceRole !== 'viewer';
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
  const bypass = useHasBusinessRoleBypass();
  const role = useCurrentProjectRole(projectId);
  if (bypass) return true;
  return role !== undefined && canDelete(role);
}

/** Whether the current user can delete resources in a project (project admins only). */
export function useCanDeleteResource(projectId?: string): boolean {
  return useCanAdministerProject(projectId);
}
