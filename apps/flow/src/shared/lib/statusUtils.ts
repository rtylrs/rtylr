import type { Task } from '../stores/flowStore';

export type TaskStatus = Task['status'];

const STATUS_ORDER: Record<string, number> = {
  backlog: 0,
  todo: 1,
  'in-progress': 2,
  in_progress: 2,
  'in-review': 3,
  in_review: 3,
  blocked: 4,
  done: 5,
  cancelled: 6,
};

/**
 * Returns a numeric order value for a task status.
 * Lower value = earlier in workflow.
 * Unknown statuses default to 99.
 */
export function statusOrder(status: string): number {
  return STATUS_ORDER[status] ?? 99;
}

/** Returns true when the task status represents completed work. */
export function isTerminalStatus(status: string): boolean {
  return status === 'done' || status === 'cancelled';
}

/** Returns true when the task is actively being worked on. */
export function isActiveStatus(status: string): boolean {
  return (
    status === 'in-progress' ||
    status === 'in-review' ||
    status === 'in_progress' ||
    status === 'in_review'
  );
}

/** Filters tasks to only those in active statuses. */
export function filterActiveTasks(tasks: Task[]): Task[] {
  return tasks.filter((t) => isActiveStatus(t.status));
}

/** Filters tasks to only those not yet completed or cancelled. */
export function filterOpenTasks(tasks: Task[]): Task[] {
  return tasks.filter((t) => !isTerminalStatus(t.status));
}

/** Returns a display label for a task status. */
export function statusLabel(status: string): string {
  const labels: Record<string, string> = {
    backlog: 'Backlog',
    todo: 'To Do',
    'in-progress': 'In Progress',
    in_progress: 'In Progress',
    'in-review': 'In Review',
    in_review: 'In Review',
    blocked: 'Blocked',
    done: 'Done',
    cancelled: 'Cancelled',
  };
  return labels[status] ?? status;
}
