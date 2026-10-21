// Board column utilities for the Flow kanban view.
import type { Task } from '../stores/flowStore';

export type TaskStatus = Task['status'];

const BOARD_STATUSES: TaskStatus[] = [
  'backlog',
  'todo',
  'in_progress',
  'in_review',
  'done',
  'cancelled',
];

const STATUS_LABELS: Record<TaskStatus, string> = {
  backlog: 'Backlog',
  todo: 'To Do',
  in_progress: 'In Progress',
  in_review: 'In Review',
  done: 'Done',
  cancelled: 'Cancelled',
};

export interface BoardColumn {
  status: TaskStatus;
  label: string;
  tasks: Task[];
}

/**
 * Builds a board column for each known status, populating tasks from
 * the provided list.
 */
export function buildBoardColumns(tasks: Task[]): BoardColumn[] {
  const byStatus = new Map<TaskStatus, Task[]>();
  for (const status of BOARD_STATUSES) {
    byStatus.set(status, []);
  }
  for (const task of tasks) {
    const bucket = byStatus.get(task.status as TaskStatus);
    if (bucket) {
      bucket.push(task);
    }
  }
  return BOARD_STATUSES.map((status) => ({
    status,
    label: STATUS_LABELS[status],
    tasks: byStatus.get(status) ?? [],
  }));
}

/**
 * Filters out columns that contain no tasks.
 */
export function nonEmptyColumns(columns: BoardColumn[]): BoardColumn[] {
  return columns.filter((c) => c.tasks.length > 0);
}

/**
 * Returns the total number of tasks across all columns.
 */
export function boardTaskCount(columns: BoardColumn[]): number {
  return columns.reduce((sum, c) => sum + c.tasks.length, 0);
}

/**
 * Returns columns sorted by canonical status order (backlog first, cancelled last).
 */
export function sortColumns(columns: BoardColumn[]): BoardColumn[] {
  return [...columns].sort(
    (a, b) => BOARD_STATUSES.indexOf(a.status) - BOARD_STATUSES.indexOf(b.status)
  );
}

/**
 * Returns the column matching the given status, or undefined.
 */
export function columnForStatus(
  columns: BoardColumn[],
  status: TaskStatus
): BoardColumn | undefined {
  return columns.find((c) => c.status === status);
}

/**
 * Returns the human-readable label for a task status.
 */
export function columnLabel(status: TaskStatus): string {
  return STATUS_LABELS[status] ?? status;
}

/**
 * How a board partitions each column into horizontal lanes. `none` is a plain
 * kanban board.
 *
 * Defined here rather than in either component: the board renders the settings
 * panel, so the panel cannot import the type from the board without a cycle, and
 * two local copies typecheck against each other only while they happen to be
 * identical. Adding a value to one and not the other would compile.
 */
export type BoardGroupKey = 'none' | 'assignee' | 'priority' | 'due';
