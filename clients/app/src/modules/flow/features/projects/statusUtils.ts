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

/**
 * The project's full status order after moving the status at `from` to `to`,
 * or null when nothing would move. The server takes the whole set at once.
 */
export function movedStatusIds(
  ordered: readonly { id: string }[],
  from: number,
  to: number
): string[] | null {
  if (from === to || from < 0 || to < 0 || from >= ordered.length || to >= ordered.length) {
    return null;
  }
  const ids = ordered.map((s) => s.id);
  const [moved] = ids.splice(from, 1);
  ids.splice(to, 0, moved);
  return ids;
}

/** Where a new status goes: into a gap, or appended then placed by one reorder. */
export interface InsertionPlan {
  /** Sent with the create when the neighbours leave room for it. */
  sortOrder?: number;
  /** Set when there is no room: the full current order and the new status's index in it. */
  placeAmong?: { ids: string[]; index: number };
}

/**
 * Where a new status inserted at `index` sorts. Uses the gap between
 * neighbours when there is one; otherwise the create appends and a single
 * reorder puts it in place.
 */
export function insertionPlan(ordered: readonly ApiStatus[], index: number): InsertionPlan {
  const at = Math.max(0, Math.min(index, ordered.length));
  const prev = ordered[at - 1];
  const next = ordered[at];
  if (!next) return { sortOrder: (prev?.sortOrder ?? -SORT_STEP) + SORT_STEP };
  const lower = prev ? prev.sortOrder : -1;
  if (next.sortOrder - lower >= 2) {
    return { sortOrder: Math.max(0, Math.floor((lower + next.sortOrder) / 2)) };
  }
  return { placeAmong: { ids: ordered.map((s) => s.id), index: at } };
}

/** Statuses rearranged to `ids`, with the sort values the server will write. */
export function applyStatusOrder(
  statuses: readonly ApiStatus[],
  ids: readonly string[]
): ApiStatus[] {
  const byId = new Map(statuses.map((s) => [s.id, s]));
  return ids.flatMap((id, i) => {
    const status = byId.get(id);
    return status ? [{ ...status, sortOrder: i * SORT_STEP }] : [];
  });
}

// Default statuses are stored as English text. Until a project renames one,
// show it in the reader's language; renamed statuses show as typed.
const BUILT_IN_STATUS_KEYS: Record<string, string> = {
  backlog: 'flowStatus.backlog',
  'to do': 'flowStatus.todo',
  todo: 'flowStatus.todo',
  'in progress': 'flowStatus.inProgress',
  'in review': 'flowStatus.inReview',
  done: 'flowStatus.done',
  cancelled: 'flowStatus.cancelled',
};

export function displayStatusName(name: string, t: (key: string) => string): string {
  const key = BUILT_IN_STATUS_KEYS[name.trim().toLowerCase()];
  return key ? t(key) : name;
}
