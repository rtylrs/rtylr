import type { ActivityEntry } from '../stores/flowStore';

/** Returns activity entries for a specific task. */
export function activityForTask(activity: ActivityEntry[], taskId: string): ActivityEntry[] {
  return activity.filter((e) => e.taskId === taskId);
}

/** Sort activity entries newest first. */
export function sortActivityDesc(activity: ActivityEntry[]): ActivityEntry[] {
  return [...activity].sort(
    (a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
  );
}

/** Sort activity entries oldest first. */
export function sortActivityAsc(activity: ActivityEntry[]): ActivityEntry[] {
  return [...activity].sort(
    (a, b) => new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime()
  );
}

/** Filter activity by event type. */
export function activityByType(
  activity: ActivityEntry[],
  type: ActivityEntry['type']
): ActivityEntry[] {
  return activity.filter((e) => e.type === type);
}

/** Returns a human-friendly description for an activity entry type. */
export function activityTypeLabel(type: ActivityEntry['type']): string {
  const labels: Record<ActivityEntry['type'], string> = {
    created: 'created this task',
    status_change: 'changed status',
    assignee_change: 'changed assignee',
    priority_change: 'changed priority',
    due_date_change: 'changed due date',
    comment: 'left a comment',
    label_added: 'added a label',
    attachment: 'added an attachment',
  };
  return labels[type] ?? type;
}

/** Returns the most recent activity for a task, or undefined. */
export function latestActivity(
  activity: ActivityEntry[],
  taskId: string
): ActivityEntry | undefined {
  const entries = activityForTask(activity, taskId);
  if (entries.length === 0) return undefined;
  return entries.reduce((latest, e) =>
    new Date(e.createdAt) > new Date(latest.createdAt) ? e : latest
  );
}

/** Returns activity entries authored by a specific actor. */
export function activityByActor(activity: ActivityEntry[], actor: string): ActivityEntry[] {
  return activity.filter((e) => e.actor === actor);
}
