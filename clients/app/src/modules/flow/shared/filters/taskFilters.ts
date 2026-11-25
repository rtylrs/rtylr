import type { Task } from '../stores/flowStore';
import { addDays, toLocalIso } from '../lib/dateUtils';
import { fieldIdFromKey, fieldValueParts } from '../lib/customFields';
import { epicKeyOf } from '../lib/taskHierarchy';

/**
 * The filter model every task list in Flow shares: the board, the list, the
 * calendar, My tasks and Team tasks.
 *
 * Values inside one field are OR'd and fields are AND'd, the way Jira's quick
 * filters behave: "Ahmed or Mohammad, and High priority".
 */

/** Stands in for "nobody" in the assignee field, next to real user ids. */
export const UNASSIGNED = 'none';

/** Stands in for "no value" in a custom field, next to option and user ids. */
export const NO_FIELD_VALUE = 'none';

export const DUE_WINDOWS = ['overdue', 'today', 'next_7', 'next_30', 'no_date'] as const;
export type DueWindow = (typeof DUE_WINDOWS)[number];

/**
 * "Created, completed or updated in the last 7 days", counting today. Rolling,
 * so a link opened next week still means the last 7 days. The project
 * dashboard's cards open the list with one of these; the Filter popover does
 * not offer them, and the bar shows each as a chip that removes it.
 */
export const CHANGE_WINDOWS = ['created_7d', 'completed_7d', 'updated_7d'] as const;
export type ChangeWindow = (typeof CHANGE_WINDOWS)[number];

export const PRIORITIES: readonly Task['priority'][] = [
  'urgent',
  'high',
  'medium',
  'low',
  'no_priority',
];

export interface TaskFilters {
  /** User ids, or UNASSIGNED. */
  assignees: string[];
  priorities: Task['priority'][];
  /** Label ids. */
  labels: string[];
  due: DueWindow[];
  /** Inclusive custom range, `YYYY-MM-DD`. Either end may be open. */
  dueFrom?: string;
  dueTo?: string;
  /**
   * Recent change windows, OR'd like due windows. Optional so filter sets
   * built before it existed (saved views, stored filters) stay valid.
   */
  changed?: ChangeWindow[];
  /**
   * Status keys. On a project they are status ids (the board's columns); on a
   * cross-project page they are the built-in status names. The caller says
   * which through `FilterContext.statusOf`.
   */
  statuses: string[];
  cycles: string[];
  milestones: string[];
  projects: string[];
  /**
   * Custom field values by field id (option ids, user ids, "true", or
   * NO_FIELD_VALUE). Only the fields of the project being viewed apply.
   * Optional so a filter set saved before custom fields existed still reads.
   */
  fields?: Record<string, string[]>;
  /** Epic ids, or NO_EPIC. An epic matches its own id (see epicKeyOf). */
  epics: string[];
  /** CI states of the task's open pull requests: failing, running, passing. */
  checks: string[];
  /** Free text matched against title, identifier, description and label names. */
  query: string;
}

/** A custom field in the Filter popover, keyed `field:<id>`. */
export type CustomFilterField = `field:${string}`;

/** The fields the Filter popover can list; `query` has its own box. */
export type FilterField =
  | 'assignees'
  | 'priorities'
  | 'labels'
  | 'due'
  | 'statuses'
  | 'cycles'
  | 'milestones'
  | 'projects'
  | 'epics'
  | 'checks'
  | CustomFilterField;

export type ListField = Exclude<FilterField, 'due' | CustomFilterField>;

export const EMPTY_TASK_FILTERS: TaskFilters = {
  assignees: [],
  priorities: [],
  labels: [],
  due: [],
  dueFrom: undefined,
  dueTo: undefined,
  changed: [],
  statuses: [],
  cycles: [],
  milestones: [],
  projects: [],
  fields: {},
  epics: [],
  checks: [],
  query: '',
};

const LIST_FIELDS: ListField[] = [
  'assignees',
  'priorities',
  'labels',
  'statuses',
  'cycles',
  'milestones',
  'projects',
  'epics',
  'checks',
];

export interface FilterContext {
  /** Local today as `YYYY-MM-DD`. Passed in so tests and renders agree on one day. */
  today: string;
  /** Label id to name, so a search for "design" finds tasks labelled Design. */
  labelNames?: ReadonlyMap<string, string>;
  /** The status key a task is filtered by. Defaults to its built-in status. */
  statusOf?: (task: Task) => string;
}

/** True when the due field narrows anything. */
export function hasDueFilter(filters: TaskFilters): boolean {
  return filters.due.length > 0 || Boolean(filters.dueFrom) || Boolean(filters.dueTo);
}

/**
 * How many fields are narrowing, for the "Filter (n)" badge. The search box is
 * not counted: it shows its own text.
 */
export function activeFieldCount(filters: TaskFilters): number {
  let n = LIST_FIELDS.filter((field) => filters[field].length > 0).length;
  n += Object.values(filters.fields ?? {}).filter((values) => values.length > 0).length;
  if (hasDueFilter(filters)) n += 1;
  if ((filters.changed ?? []).length > 0) n += 1;
  return n;
}

/** The values picked for a list field or a custom field. */
export function selectedValues(
  filters: TaskFilters,
  field: ListField | CustomFilterField
): string[] {
  const fieldId = fieldIdFromKey(field);
  if (fieldId !== null) return filters.fields?.[fieldId] ?? [];
  return filters[field as ListField];
}

/** The filters with one field's picked values replaced. */
export function withSelectedValues(
  filters: TaskFilters,
  field: ListField | CustomFilterField,
  values: string[]
): TaskFilters {
  const fieldId = fieldIdFromKey(field);
  if (fieldId === null) return { ...filters, [field]: values };
  const fields = { ...(filters.fields ?? {}) };
  if (values.length > 0) fields[fieldId] = values;
  else delete fields[fieldId];
  return { ...filters, fields };
}

/** Every field cleared, the search text kept. */
export function clearFilterFields(filters: TaskFilters): TaskFilters {
  return { ...EMPTY_TASK_FILTERS, query: filters.query };
}

export function hasActiveTaskFilters(filters: TaskFilters): boolean {
  return activeFieldCount(filters) > 0 || filters.query.trim() !== '';
}

/**
 * Lower-case and drop combining marks, so "cafe" finds "Café" and an Arabic
 * search ignores harakat.
 */
export function normalizeSearchText(value: string): string {
  return value
    .normalize('NFKD')
    .replace(/[̀-ًͯ-ٰٟ]/g, '')
    .toLowerCase();
}

function dayOf(date: string | undefined): string {
  return date ? (date.split('T')[0] ?? '') : '';
}

// Day strings, never Dates: a due date is a calendar day, and `YYYY-MM-DD`
// sorts in date order (see myTasksBuckets for the bug that taught this).
function inDueWindow(due: string, window: DueWindow, today: string): boolean {
  switch (window) {
    case 'overdue':
      return due !== '' && due < today;
    case 'today':
      return due === today;
    case 'next_7':
      return due !== '' && due >= today && due <= addDays(today, 7);
    case 'next_30':
      return due !== '' && due >= today && due <= addDays(today, 30);
    case 'no_date':
      return due === '';
  }
}

function matchesDue(task: Task, filters: TaskFilters, today: string): boolean {
  if (!hasDueFilter(filters)) return true;
  const due = dayOf(task.dueDate);
  if (filters.due.some((w) => inDueWindow(due, w, today))) return true;
  if (filters.dueFrom || filters.dueTo) {
    if (due === '') return false;
    if (filters.dueFrom && due < filters.dueFrom) return false;
    if (filters.dueTo && due > filters.dueTo) return false;
    return true;
  }
  return false;
}

/** The reader's own calendar day of a server timestamp, or '' when there is none. */
function localDayOf(stamp: string | undefined): string {
  if (!stamp) return '';
  const at = new Date(stamp);
  return Number.isNaN(at.getTime()) ? '' : toLocalIso(at);
}

function changeStamp(task: Task, window: ChangeWindow): string | undefined {
  switch (window) {
    case 'created_7d':
      return task.createdAt;
    case 'completed_7d':
      return task.completedAt;
    case 'updated_7d':
      return task.updatedAt;
  }
}

function matchesChanged(task: Task, filters: TaskFilters, today: string): boolean {
  const windows = filters.changed ?? [];
  if (windows.length === 0) return true;
  const from = addDays(today, -6);
  return windows.some((w) => {
    const day = localDayOf(changeStamp(task, w));
    return day !== '' && day >= from && day <= today;
  });
}

function matchesQuery(task: Task, query: string, ctx: FilterContext): boolean {
  const tokens = normalizeSearchText(query).split(/\s+/).filter(Boolean);
  if (tokens.length === 0) return true;
  const labelText = (task.labelIds ?? []).map((id) => ctx.labelNames?.get(id) ?? '').join(' ');
  const haystack = normalizeSearchText(
    [task.identifier, task.title, task.description ?? '', labelText].join('\n')
  );
  return tokens.every((token) => haystack.includes(token));
}

/** Whether one task passes every active field. */
export function matchesTaskFilters(task: Task, filters: TaskFilters, ctx: FilterContext): boolean {
  if (filters.assignees.length > 0) {
    const key = task.assigneeId ? task.assigneeId : UNASSIGNED;
    if (!filters.assignees.includes(key)) return false;
  }
  if (filters.priorities.length > 0 && !filters.priorities.includes(task.priority)) return false;
  if (filters.labels.length > 0) {
    const own = task.labelIds ?? [];
    if (!filters.labels.some((id) => own.includes(id))) return false;
  }
  if (filters.statuses.length > 0) {
    const status = ctx.statusOf ? ctx.statusOf(task) : task.status;
    if (!filters.statuses.includes(status)) return false;
  }
  if (filters.cycles.length > 0 && !filters.cycles.includes(task.cycleId ?? '')) return false;
  if (filters.milestones.length > 0 && !filters.milestones.includes(task.milestoneId ?? '')) {
    return false;
  }
  if (filters.projects.length > 0 && !filters.projects.includes(task.projectId)) return false;
  for (const [fieldId, picked] of Object.entries(filters.fields ?? {})) {
    if (picked.length === 0) continue;
    const parts = fieldValueParts(task.fieldValues?.[fieldId]);
    const hit = picked.some((v) => (v === NO_FIELD_VALUE ? parts.length === 0 : parts.includes(v)));
    if (!hit) return false;
  }
  if (filters.epics.length > 0 && !filters.epics.includes(epicKeyOf(task))) return false;
  if (filters.checks.length > 0 && !filters.checks.includes(task.prChecksState ?? '')) return false;
  if (!matchesDue(task, filters, ctx.today)) return false;
  if (!matchesChanged(task, filters, ctx.today)) return false;
  return matchesQuery(task, filters.query, ctx);
}

/** Adds the value when absent and removes it when present. */
export function toggleValue<T extends string>(values: readonly T[], value: T): T[] {
  return values.includes(value) ? values.filter((v) => v !== value) : [...values, value];
}

/**
 * "Only mine" is not a separate flag: it is the assignee field holding exactly
 * the current user, so the avatar strip and the pill can never disagree.
 */
export function isOnlyMine(filters: TaskFilters, userId: string | null): boolean {
  return userId !== null && filters.assignees.length === 1 && filters.assignees[0] === userId;
}

export function toggleOnlyMine(filters: TaskFilters, userId: string | null): TaskFilters {
  if (userId === null) return filters;
  return { ...filters, assignees: isOnlyMine(filters, userId) ? [] : [userId] };
}

/**
 * The URL and storage shape: the same fields, with empty ones left out so a
 * shared link stays short.
 */
export type SerializedTaskFilters = Partial<TaskFilters>;

export function serializeTaskFilters(filters: TaskFilters): SerializedTaskFilters | undefined {
  const out: Record<string, unknown> = {};
  for (const field of LIST_FIELDS) {
    if (filters[field].length > 0) out[field] = filters[field];
  }
  const fields = Object.fromEntries(
    Object.entries(filters.fields ?? {}).filter(([, values]) => values.length > 0)
  );
  if (Object.keys(fields).length > 0) out.fields = fields;
  if (filters.due.length > 0) out.due = filters.due;
  if (filters.dueFrom) out.dueFrom = filters.dueFrom;
  if (filters.dueTo) out.dueTo = filters.dueTo;
  if (filters.changed && filters.changed.length > 0) out.changed = filters.changed;
  if (filters.query.trim() !== '') out.query = filters.query;
  return Object.keys(out).length > 0 ? (out as SerializedTaskFilters) : undefined;
}

const DAY = /^\d{4}-\d{2}-\d{2}$/;

function stringList(value: unknown): string[] {
  if (!Array.isArray(value)) return [];
  return value.filter((v) => typeof v === 'string' || typeof v === 'number').map((v) => String(v));
}

function fieldLists(value: unknown): Record<string, string[]> {
  if (typeof value !== 'object' || value === null || Array.isArray(value)) return {};
  const out: Record<string, string[]> = {};
  for (const [id, values] of Object.entries(value as Record<string, unknown>)) {
    const list = stringList(values);
    if (/^\d+$/.test(id) && list.length > 0) out[id] = list;
  }
  return out;
}

/**
 * Accepts anything (a hand-edited URL, an old localStorage entry) and returns
 * a complete, valid filter set. Unknown values are dropped rather than failing
 * the whole set.
 */
export function parseTaskFilters(value: unknown): TaskFilters {
  if (typeof value !== 'object' || value === null) return EMPTY_TASK_FILTERS;
  const raw = value as Record<string, unknown>;
  const day = (v: unknown): string | undefined =>
    typeof v === 'string' && DAY.test(v) ? v : undefined;
  return {
    assignees: stringList(raw.assignees),
    priorities: stringList(raw.priorities).filter((p): p is Task['priority'] =>
      (PRIORITIES as readonly string[]).includes(p)
    ),
    labels: stringList(raw.labels),
    due: stringList(raw.due).filter((d): d is DueWindow =>
      (DUE_WINDOWS as readonly string[]).includes(d)
    ),
    dueFrom: day(raw.dueFrom),
    dueTo: day(raw.dueTo),
    changed: stringList(raw.changed).filter((c): c is ChangeWindow =>
      (CHANGE_WINDOWS as readonly string[]).includes(c)
    ),
    statuses: stringList(raw.statuses),
    cycles: stringList(raw.cycles),
    milestones: stringList(raw.milestones),
    projects: stringList(raw.projects),
    fields: fieldLists(raw.fields),
    epics: stringList(raw.epics),
    checks: stringList(raw.checks),
    query: typeof raw.query === 'string' ? raw.query : '',
  };
}
