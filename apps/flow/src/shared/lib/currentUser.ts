import { useUserStore } from '../stores/userStore';

/** Fallback user id when no authenticated user is available */
export const DEMO_USER_ID = 'u1';

export function getCurrentUserId(): string {
  return useUserStore.getState().user?.id ?? DEMO_USER_ID;
}

/** Whether the task counts as "mine" for My tasks. */
export function isTaskMine(taskAssigneeId: string): boolean {
  return taskAssigneeId === getCurrentUserId();
}
