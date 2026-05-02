import { useUserStore } from '../stores/userStore';

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
