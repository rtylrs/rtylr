import type { Task } from '../../../shared/stores/flowStore';
import type { ApiStatus } from '../../../shared/services/flowApi';
import { PRIORITY_ORDER } from '../taskGroups';
import { defaultStatus, sortStatuses, statusIdOf } from '../statusUtils';

/**
 * The List view's columns. Custom fields are not columns yet: they need the
 * custom field values on the task list response first (T-579).
 */
export type ListColumnId =
  | 'priority'
  | 'title'
  | 'status'
  | 'assignee'
  | 'startDate'
  | 'dueDate'
  | 'estimate'
  | 'labels'
  | 'key'
  | 'created';

export interface ListColumnDef {
  id: ListColumnId;
  labelKey: string;
  /**
   * Width in px. The title column fills whatever room is left, so for it this
   * is the narrowest it gets rather than a fixed width.
   */
  width: number;
  minWidth: number;
  /** Whether the cell can be edited in place by someone who may edit the task. */
  editable: boolean;
  /** Dates and numbers line up at the end of the cell. */
  alignEnd?: boolean;
}

export const LIST_COLUMNS: readonly ListColumnDef[] = [
  {
    id: 'priority',
    labelKey: 'flow.projectList.column.priority',
    width: 16,
    minWidth: 16,
    editable: true,
  },
  {
    id: 'title',
    labelKey: 'flow.projectList.columnTitle',
    width: 160,
    minWidth: 120,
    editable: false,
  },
  {
    id: 'status',
    labelKey: 'flow.projectList.column.status',
    width: 140,
    minWidth: 80,
    editable: true,
  },
  {
    id: 'assignee',
    labelKey: 'flow.projectList.column.assignee',
    width: 20,
    minWidth: 20,
    editable: true,
  },
  {
    id: 'startDate',
    labelKey: 'flow.projectList.column.startDate',
    width: 96,
    minWidth: 72,
    editable: true,
    alignEnd: true,
  },
  {
    id: 'dueDate',
    labelKey: 'flowCommon.dueDate',
    width: 96,
    minWidth: 72,
    editable: true,
    alignEnd: true,
  },
  {
    id: 'estimate',
    labelKey: 'flow.projectList.column.estimate',
    width: 72,
    minWidth: 56,
    editable: true,
    alignEnd: true,
  },
  {
    id: 'labels',
    labelKey: 'flow.projectList.column.labels',
    width: 160,
    minWidth: 80,
    editable: true,
  },
  { id: 'key', labelKey: 'flow.projectList.column.key', width: 80, minWidth: 56, editable: false },
  {
    id: 'created',
    labelKey: 'flow.projectList.column.created',
    width: 96,
    minWidth: 72,
    editable: false,
    alignEnd: true,
  },
];

const COLUMN_BY_ID = new Map(LIST_COLUMNS.map((column) => [column.id, column]));

export function listColumn(id: ListColumnId): ListColumnDef {
  return COLUMN_BY_ID.get(id)!;
}

/** Exactly what the List showed before columns could be chosen. */
export const DEFAULT_LIST_COLUMNS: readonly ListColumnId[] = [
  'priority',
  'title',
  'assignee',
  'dueDate',
];

export const MAX_COLUMN_WIDTH = 640;

export type SortDirection = 'asc' | 'desc';

export interface ListSort {
  column: ListColumnId;
  direction: SortDirection;
}

export interface ListColumnPrefs {
  /** Shown columns, in order. Always holds the title. */
  columns: ListColumnId[];
  /** Widths the user dragged to; a column without one uses its default. */
  widths: Partial<Record<ListColumnId, number>>;
  /** Null keeps the server's order. */
  sort: ListSort | null;
}

export const DEFAULT_LIST_PREFS: ListColumnPrefs = {
  columns: [...DEFAULT_LIST_COLUMNS],
  widths: {},
  sort: null,
};

const isColumnId = (value: unknown): value is ListColumnId =>
  typeof value === 'string' && COLUMN_BY_ID.has(value as ListColumnId);

export function clampColumnWidth(id: ListColumnId, width: number): number {
  const { minWidth } = listColumn(id);
  if (!Number.isFinite(width)) return listColumn(id).width;
  return Math.round(Math.min(MAX_COLUMN_WIDTH, Math.max(minWidth, width)));
}

export function columnWidth(prefs: ListColumnPrefs, id: ListColumnId): number {
  return prefs.widths[id] ?? listColumn(id).width;
}

/** Row chrome around the columns: inline padding, the Done box, the More button, the gaps. */
const ROW_PADDING = 32;
const DONE_CELL = 16;
const MORE_CELL = 20;
const CELL_GAP = 12;

/**
 * The narrowest a row can get before the List scrolls sideways: every column
 * at its width (the title at its minimum) plus the row's own chrome.
 */
export function rowMinWidth(prefs: ListColumnPrefs): number {
  const columns = prefs.columns.reduce((sum, id) => sum + columnWidth(prefs, id), 0);
  const cells = prefs.columns.length + 2;
  return ROW_PADDING + DONE_CELL + MORE_CELL + columns + CELL_GAP * (cells - 1);
}

/**
 * Reads stored preferences defensively: they outlive releases, so a column
 * that no longer exists, a duplicate, a missing title or a width out of range
 * is repaired rather than trusted.
 */
export function normalizeListPrefs(raw: unknown): ListColumnPrefs {
  if (!raw || typeof raw !== 'object') return { ...DEFAULT_LIST_PREFS, widths: {} };
  const value = raw as Record<string, unknown>;

  const columns: ListColumnId[] = [];
  if (Array.isArray(value.columns)) {
    for (const id of value.columns) {
      if (isColumnId(id) && !columns.includes(id)) columns.push(id);
    }
  }
  if (columns.length === 0) columns.push(...DEFAULT_LIST_COLUMNS);
  if (!columns.includes('title')) columns.unshift('title');

  const widths: Partial<Record<ListColumnId, number>> = {};
  if (value.widths && typeof value.widths === 'object') {
    for (const [id, width] of Object.entries(value.widths as Record<string, unknown>)) {
      if (isColumnId(id) && typeof width === 'number' && Number.isFinite(width)) {
        widths[id] = clampColumnWidth(id, width);
      }
    }
  }

  let sort: ListSort | null = null;
  const rawSort = value.sort as Record<string, unknown> | null | undefined;
  if (
    rawSort &&
    isColumnId(rawSort.column) &&
    (rawSort.direction === 'asc' || rawSort.direction === 'desc')
  ) {
    sort = { column: rawSort.column, direction: rawSort.direction };
  }

  return { columns, widths, sort };
}

/**
 * Per user and per project, in this browser only: there is no server side
 * store for view preferences yet.
 */
export const listPrefsStorageKey = (userId: string, projectId: string): string =>
  `flow:list:columns:${userId}:${projectId}`;

export function loadListPrefs(userId: string | null, projectId: string): ListColumnPrefs {
  if (!userId || !projectId) return normalizeListPrefs(null);
  try {
    const raw = localStorage.getItem(listPrefsStorageKey(userId, projectId));
    return normalizeListPrefs(raw ? JSON.parse(raw) : null);
  } catch {
    // Storage blocked or the value is not JSON: start from the defaults.
    return normalizeListPrefs(null);
  }
}

export function saveListPrefs(
  userId: string | null,
  projectId: string,
  prefs: ListColumnPrefs
): void {
  if (!userId || !projectId) return;
  try {
    localStorage.setItem(listPrefsStorageKey(userId, projectId), JSON.stringify(prefs));
  } catch {
    // A preference that cannot be stored still applies for this visit.
  }
}

/** Shows or hides a column. The title cannot be hidden. */
export function toggleListColumn(prefs: ListColumnPrefs, id: ListColumnId): ListColumnPrefs {
  if (id === 'title') return prefs;
  if (prefs.columns.includes(id)) {
    return {
      ...prefs,
      columns: prefs.columns.filter((column) => column !== id),
      sort: prefs.sort?.column === id ? null : prefs.sort,
    };
  }
  return { ...prefs, columns: [...prefs.columns, id] };
}

/** Moves a shown column one place earlier (-1) or later (+1). */
export function moveListColumn(
  prefs: ListColumnPrefs,
  id: ListColumnId,
  delta: -1 | 1
): ListColumnPrefs {
  const from = prefs.columns.indexOf(id);
  const to = from + delta;
  if (from < 0 || to < 0 || to >= prefs.columns.length) return prefs;
  const columns = [...prefs.columns];
  [columns[from], columns[to]] = [columns[to], columns[from]];
  return { ...prefs, columns };
}

/** Header click: ascending, then descending, then back to the server order. */
export function nextSort(current: ListSort | null, column: ListColumnId): ListSort | null {
  if (current?.column !== column) return { column, direction: 'asc' };
  if (current.direction === 'asc') return { column, direction: 'desc' };
  return null;
}

/**
 * The status a row shows: its own when the project has it, else the status
 * with the same built-in value, else the project default. Matches where
 * buildListGroups files the task.
 */
export function resolveTaskStatus(
  task: Pick<Task, 'statusId' | 'sectionId' | 'status'>,
  ordered: readonly ApiStatus[]
): ApiStatus | undefined {
  const own = statusIdOf(task);
  return (
    (own ? ordered.find((s) => s.id === own) : undefined) ??
    ordered.find((s) => s.legacyStatus === task.status) ??
    defaultStatus(ordered)
  );
}

export interface ListSortContext {
  statuses: readonly ApiStatus[];
  /** Label names by id, for sorting by the labels a task carries. */
  labelNames: ReadonlyMap<string, string>;
}

type SortValue = string | number | null;

function sortValue(
  task: Task,
  column: ListColumnId,
  ctx: ListSortContext,
  ordered: readonly ApiStatus[]
): SortValue {
  switch (column) {
    case 'priority':
      return PRIORITY_ORDER.indexOf(task.priority);
    case 'title':
      return task.title || null;
    case 'status': {
      const status = resolveTaskStatus(task, ordered);
      return status ? ordered.indexOf(status) : null;
    }
    case 'assignee':
      return task.assignee || null;
    case 'startDate':
      return task.startDate || null;
    case 'dueDate':
      return task.dueDate || null;
    case 'estimate':
      return task.estimate != null && task.estimate > 0 ? task.estimate : null;
    case 'labels': {
      const names = (task.labelIds ?? [])
        .map((id) => ctx.labelNames.get(String(id)))
        .filter((name): name is string => !!name)
        .sort((a, b) => a.localeCompare(b));
      return names.length ? names.join(', ') : null;
    }
    case 'key':
      return task.identifier || null;
    case 'created':
      return task.createdAt || null;
  }
}

function compareValues(a: string | number, b: string | number): number {
  if (typeof a === 'number' && typeof b === 'number') return a - b;
  return String(a).localeCompare(String(b), undefined, { numeric: true, sensitivity: 'base' });
}

/**
 * Sorts a copy of the tasks by one column. Rows with no value always go last,
 * whichever way the column is sorted, and ties keep their server order.
 */
export function sortListTasks(
  tasks: readonly Task[],
  sort: ListSort | null,
  ctx: ListSortContext
): Task[] {
  if (!sort) return [...tasks];
  const ordered = sortStatuses(ctx.statuses);
  const sign = sort.direction === 'asc' ? 1 : -1;
  return tasks
    .map((task, index) => ({ task, index, value: sortValue(task, sort.column, ctx, ordered) }))
    .sort((a, b) => {
      if (a.value === null || b.value === null) {
        if (a.value === b.value) return a.index - b.index;
        return a.value === null ? 1 : -1;
      }
      return compareValues(a.value, b.value) * sign || a.index - b.index;
    })
    .map((entry) => entry.task);
}
