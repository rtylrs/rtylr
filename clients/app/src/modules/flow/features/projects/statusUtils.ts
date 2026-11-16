import type { ApiStatus } from '../../shared/services/flowApi';
import type { Task } from '../../shared/stores/flowStore';

export type StatusBucket = ApiStatus['bucket'];

export const STATUS_BUCKETS: readonly StatusBucket[] = ['backlog', 'active', 'done', 'cancelled'];

export const STATUS_COLORS: readonly string[] = [
  '#ef4444',
  '#f97316',
  '#eab308',
  '#22c55e',
  '#3b82f6',
  '#8b5cf6',
  '#ec4899',
  '#64C5B9',
];

/** Done and cancelled columns hide cards closed longer ago than this. */
export const CLOSED_WINDOW_DAYS = 14;

const DAY_MS = 86_400_000;
const SORT_STEP = 10;

export function sortStatuses(statuses: readonly ApiStatus[]): ApiStatus[] {
  return [...statuses].sort((a, b) => a.sortOrder - b.sortOrder || Number(a.id) - Number(b.id));
}

export function firstDoneStatus(statuses: readonly ApiStatus[]): ApiStatus | undefined {
  return sortStatuses(statuses).find((s) => s.bucket === 'done');
}

/** The status a new or reopened task gets: the flagged default, else the first open one. */
export function defaultStatus(statuses: readonly ApiStatus[]): ApiStatus | undefined {
  const ordered = sortStatuses(statuses);
  return (
    ordered.find((s) => s.isDefault) ?? ordered.find((s) => !isClosedBucket(s.bucket)) ?? ordered[0]
  );
}

/** The server refuses to delete a project's last done status. */
export function isOnlyDoneStatus(status: ApiStatus, statuses: readonly ApiStatus[]): boolean {
  return (
    status.bucket === 'done' && !statuses.some((s) => s.id !== status.id && s.bucket === 'done')
  );
}

export function isClosedBucket(bucket: StatusBucket): boolean {
  return bucket === 'done' || bucket === 'cancelled';
}

/** The status a task sits on. section_id mirrors status_id, so either names it. */
export function statusIdOf(task: Pick<Task, 'statusId' | 'sectionId'>): string | undefined {
  return task.statusId || task.sectionId || undefined;
}

/**
 * Whether a card in a done or cancelled column is recent enough to show by
 * default. Done reads completedAt, falling back to updatedAt for rows closed
 * before completed_at was stamped; cancelled reads updatedAt. A task with no
 * usable timestamp shows rather than silently disappearing.
 */
export function closedRecently(
  task: Pick<Task, 'completedAt' | 'updatedAt'>,
  bucket: StatusBucket,
  now: number,
  days: number = CLOSED_WINDOW_DAYS
): boolean {
  const stamp = bucket === 'done' ? (task.completedAt ?? task.updatedAt) : task.updatedAt;
  const at = stamp ? Date.parse(stamp) : Number.NaN;
  if (Number.isNaN(at)) return true;
  return now - at <= days * DAY_MS;
}

export interface SortUpdate {
  status: ApiStatus;
  sortOrder: number;
}

/**
 * The writes that move the status at `from` to `to`. Existing sort values are
 * reused in the new order, so a swap of neighbours is two writes; ties are
 * renumbered first because swapping equal values would change nothing.
 */
export function reorderUpdates(
  ordered: readonly ApiStatus[],
  from: number,
  to: number
): SortUpdate[] {
  if (from === to || from < 0 || to < 0 || from >= ordered.length || to >= ordered.length) {
    return [];
  }
  const next = [...ordered];
  const [moved] = next.splice(from, 1);
  next.splice(to, 0, moved);
  const values = ordered.map((s) => s.sortOrder);
  const distinct = new Set(values).size === values.length;
  return next.flatMap((status, i) => {
    const sortOrder = distinct ? values[i] : i * SORT_STEP;
    return sortOrder === status.sortOrder ? [] : [{ status, sortOrder }];
  });
}

/**
 * Where a new status inserted at `index` sorts, and any writes needed to make
 * room. Uses the gap between neighbours when there is one; otherwise every
 * status is renumbered in steps of ten.
 */
export function insertionPlan(
  ordered: readonly ApiStatus[],
  index: number
): { sortOrder: number; updates: SortUpdate[] } {
  const at = Math.max(0, Math.min(index, ordered.length));
  const prev = ordered[at - 1];
  const next = ordered[at];
  if (!next) return { sortOrder: (prev?.sortOrder ?? -SORT_STEP) + SORT_STEP, updates: [] };
  const lower = prev ? prev.sortOrder : -1;
  if (next.sortOrder - lower >= 2) {
    return { sortOrder: Math.max(0, Math.floor((lower + next.sortOrder) / 2)), updates: [] };
  }
  const updates = ordered.flatMap((status, i) => {
    const sortOrder = (i < at ? i : i + 1) * SORT_STEP;
    return sortOrder === status.sortOrder ? [] : [{ status, sortOrder }];
  });
  return { sortOrder: at * SORT_STEP, updates };
}
