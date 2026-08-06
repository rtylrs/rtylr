import { useUserStore } from '../stores/userStore';
import { useFlowStore, type Member } from '../stores/flowStore';
import { findMemberById } from './memberUtils';
import { canEditTask, canDelete } from './roleUtils';

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

/**
 * The signed-in user's Flow role, reactive to store updates.
 * flowStore.members isn't actually workspace-scoped today - it's seeded
 * once from the business team (see FlowLayout.tsx), so this is the same
 * role everywhere in the app, not a per-workspace lookup.
 * Returns undefined while members haven't loaded yet or the user isn't
 * found - callers should treat that as "not permitted" (fail closed).
 */
export function useCurrentMemberRole(): Member['role'] | undefined {
  const userId = useUserStore((s) => s.user?.id) ?? (import.meta.env.PROD ? null : DEMO_USER_ID);
  const members = useFlowStore((s) => s.members);
  return userId ? findMemberById(members, userId)?.role : undefined;
}

/** Whether the current user can edit task fields (title, status, assignee, etc). */
export function useCanEditTask(): boolean {
  const role = useCurrentMemberRole();
  return role !== undefined && canEditTask(role);
}

/** Whether the current user can delete resources (owner/admin only). */
export function useCanDeleteResource(): boolean {
  const role = useCurrentMemberRole();
  return role !== undefined && canDelete(role);
}
