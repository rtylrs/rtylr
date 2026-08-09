import type { Task } from '../stores/flowStore';

export type Priority = Task['priority'];

const PRIORITY_WEIGHTS: Record<string, number> = {
  urgent: 4,
  high: 3,
  medium: 2,
  low: 1,
};

/**
 * Returns a numeric weight for a task priority.
 * Higher weight = higher priority.
 * Unknown priorities default to 0.
 */
export function priorityWeight(priority: string): number {
  return PRIORITY_WEIGHTS[priority] ?? 0;
}

/** Returns true when the priority is considered high-urgency (urgent or high). */
export function isHighPriority(priority: string): boolean {
  return priority === 'urgent' || priority === 'high';
}

/** Filters tasks to only those with high or urgent priority. */
export function filterHighPriorityTasks(tasks: Task[]): Task[] {
  return tasks.filter((t) => isHighPriority(t.priority));
}

/** Returns a human-readable label for a priority value. */
export function priorityLabel(priority: string): string {
  const labels: Record<string, string> = {
    urgent: 'Urgent',
    high: 'High',
    medium: 'Medium',
    low: 'Low',
  };
  return labels[priority] ?? priority;
}
