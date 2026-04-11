import { isFlowLocal } from './flowLocalMode';
import { useUserStore } from '../stores/userStore';

/** Demo assignee id used in local seed data */
export const DEMO_USER_ID = 'u1';

export function getCurrentUserId(): string {
  return useUserStore.getState().user?.id ?? DEMO_USER_ID;
}

/** Whether the task counts as "mine" for My tasks. In local mode, seeded u1 tasks are included for any user. */
export function isTaskMine(taskAssigneeId: string): boolean {
  const uid = getCurrentUserId();
  if (taskAssigneeId === uid) return true;
  if (isFlowLocal() && taskAssigneeId === DEMO_USER_ID) return true;
  return false;
}
