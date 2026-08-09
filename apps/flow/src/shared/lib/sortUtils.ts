import type { Task } from '../stores/flowStore';
import { priorityWeight } from './priorityUtils';
import { statusOrder } from './statusUtils';

export type TaskSortKey = 'priority' | 'status' | 'dueDate' | 'created' | 'updated' | 'title';
export type SortDir = 'asc' | 'desc';

// Priority asc = highest priority first (urgent before low).
// Priority desc = lowest priority first (low before urgent).
function byPriority(a: Task, b: Task, dir: SortDir): number {
  const diff = priorityWeight(b.priority) - priorityWeight(a.priority);
  return dir === 'asc' ? diff : -diff;
}

// Status asc = earliest pipeline stage first (backlog before done).
function byStatus(a: Task, b: Task, dir: SortDir): number {
  const diff = statusOrder(a.status) - statusOrder(b.status);
  return dir === 'asc' ? diff : -diff;
}

function byDueDate(a: Task, b: Task, dir: SortDir): number {
  if (!a.dueDate && !b.dueDate) return 0;
  if (!a.dueDate) return 1;
  if (!b.dueDate) return -1;
  const diff = a.dueDate.localeCompare(b.dueDate);
  return dir === 'asc' ? diff : -diff;
}

function byCreated(a: Task, b: Task, dir: SortDir): number {
  const diff = a.createdAt.localeCompare(b.createdAt);
  return dir === 'asc' ? diff : -diff;
}

function byUpdated(a: Task, b: Task, dir: SortDir): number {
  const diff = a.updatedAt.localeCompare(b.updatedAt);
  return dir === 'asc' ? diff : -diff;
}

function byTitle(a: Task, b: Task, dir: SortDir): number {
  const diff = a.title.localeCompare(b.title);
  return dir === 'asc' ? diff : -diff;
}

const COMPARATORS: Record<TaskSortKey, (a: Task, b: Task, dir: SortDir) => number> = {
  priority: byPriority,
  status: byStatus,
  dueDate: byDueDate,
  created: byCreated,
  updated: byUpdated,
  title: byTitle,
};

/** Sorts a list of tasks by the given key and direction. */
export function sortTasks(tasks: Task[], key: TaskSortKey, dir: SortDir = 'asc'): Task[] {
  return [...tasks].sort((a, b) => COMPARATORS[key](a, b, dir));
}

/** Sorts tasks by multiple keys in order of precedence. */
export function sortTasksMultiKey(
  tasks: Task[],
  criteria: Array<{ key: TaskSortKey; dir: SortDir }>
): Task[] {
  return [...tasks].sort((a, b) => {
    for (const { key, dir } of criteria) {
      const diff = COMPARATORS[key](a, b, dir);
      if (diff !== 0) return diff;
    }
    return 0;
  });
}
