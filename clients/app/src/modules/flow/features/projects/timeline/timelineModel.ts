/**
 * The project Timeline's layout (T-342): which days to show, where each task's
 * bar sits, how rows group, and which bars a dependency arrow joins. Kept free
 * of React so it can be tested without a DOM.
 */

export interface TimelineTask {
  id: string;
  title: string;
  identifier?: string;
  startDate?: string;
  dueDate?: string;
  done: boolean;
  assigneeId?: string;
  assigneeName?: string;
  milestoneId?: string;
  /** Ids of the tasks that block this one. */
  dependencies?: string[];
}

export type TimelineGroupBy = 'none' | 'assignee' | 'milestone';

export interface TimelineBar {
  task: TimelineTask;
  start: string;
  end: string;
  /** Days from the window's first day to the bar's first day. */
  offset: number;
  /** Days the bar covers, at least 1. */
  span: number;
}

export interface TimelineGroup {
  key: string;
  label: string;
  bars: TimelineBar[];
}

export type TimelineRow =
  { kind: 'group'; key: string; label: string; count: number } | { kind: 'task'; bar: TimelineBar };

export interface TimelineArrow {
  fromTaskId: string;
  toTaskId: string;
  fromRow: number;
  toRow: number;
  /** Day index where the blocker ends (exclusive) and the blocked task starts. */
  fromDay: number;
  toDay: number;
}

const DAY_MS = 24 * 60 * 60 * 1000;
const ISO_DAY = /^\d{4}-\d{2}-\d{2}/;
const MIN_DAYS = 28;
const MAX_DAYS = 366;

const dayOf = (value?: string): string | undefined =>
  value && ISO_DAY.test(value) ? value.slice(0, 10) : undefined;

function utc(day: string): number {
  return Date.UTC(Number(day.slice(0, 4)), Number(day.slice(5, 7)) - 1, Number(day.slice(8, 10)));
}

export function daysBetween(from: string, to: string): number {
  return Math.round((utc(to) - utc(from)) / DAY_MS);
}

export function addDayCount(day: string, count: number): string {
  const y = Number(day.slice(0, 4));
  const m = Number(day.slice(5, 7)) - 1;
  return new Date(Date.UTC(y, m, Number(day.slice(8, 10)) + count)).toISOString().slice(0, 10);
}

/** The bar's first and last day: start to due, or the one date it has. */
export function taskSpan(task: TimelineTask): { start: string; end: string } | null {
  const start = dayOf(task.startDate);
  const due = dayOf(task.dueDate);
  const first = start ?? due;
  const last = due ?? start;
  if (!first || !last) return null;
  return first <= last ? { start: first, end: last } : { start: last, end: first };
}

/**
 * The days on screen: a few before the earliest bar (or today) to two weeks
 * past the latest, at least four weeks and at most a year.
 */
export function timelineWindow(tasks: readonly TimelineTask[], today: string): string[] {
  let first = addDayCount(today, -3);
  let last = addDayCount(today, 14);
  for (const task of tasks) {
    const span = taskSpan(task);
    if (!span) continue;
    if (span.start < first) first = addDayCount(span.start, -2);
    if (span.end > last) last = addDayCount(span.end, 7);
  }
  let count = Math.min(Math.max(daysBetween(first, last) + 1, MIN_DAYS), MAX_DAYS);
  if (daysBetween(first, today) >= MAX_DAYS) first = addDayCount(today, -7);
  count = Math.max(count, 1);
  return Array.from({ length: count }, (_, i) => addDayCount(first, i));
}

export function buildBars(tasks: readonly TimelineTask[], windowStart: string): TimelineBar[] {
  const bars: TimelineBar[] = [];
  for (const task of tasks) {
    const span = taskSpan(task);
    if (!span) continue;
    bars.push({
      task,
      start: span.start,
      end: span.end,
      offset: daysBetween(windowStart, span.start),
      span: daysBetween(span.start, span.end) + 1,
    });
  }
  return bars.sort(
    (a, b) =>
      a.start.localeCompare(b.start) ||
      a.end.localeCompare(b.end) ||
      a.task.title.localeCompare(b.task.title)
  );
}

/**
 * Bars grouped by assignee or milestone; groups follow the order of the
 * labels map, and tasks with no value go last under `noneLabel`.
 */
export function groupBars(
  bars: readonly TimelineBar[],
  groupBy: TimelineGroupBy,
  labels: ReadonlyMap<string, string>,
  noneLabel: string
): TimelineGroup[] {
  if (groupBy === 'none') return [{ key: 'all', label: '', bars: [...bars] }];
  const keyOf = (bar: TimelineBar) =>
    (groupBy === 'assignee' ? bar.task.assigneeId : bar.task.milestoneId) || '';
  const byKey = new Map<string, TimelineBar[]>();
  for (const bar of bars) {
    const key = keyOf(bar);
    byKey.set(key, [...(byKey.get(key) ?? []), bar]);
  }
  const keys = [...byKey.keys()].sort((a, b) => {
    if (a === '') return 1;
    if (b === '') return -1;
    return (labels.get(a) ?? a).localeCompare(labels.get(b) ?? b);
  });
  return keys.map((key) => ({
    key: key || 'none',
    label: key === '' ? noneLabel : (labels.get(key) ?? key),
    bars: byKey.get(key) ?? [],
  }));
}

/** The rows in screen order: a header per group (when grouped), then its bars. */
export function flattenRows(groups: readonly TimelineGroup[], grouped: boolean): TimelineRow[] {
  const rows: TimelineRow[] = [];
  for (const group of groups) {
    if (grouped)
      rows.push({ kind: 'group', key: group.key, label: group.label, count: group.bars.length });
    for (const bar of group.bars) rows.push({ kind: 'task', bar });
  }
  return rows;
}

/** One arrow per "blocks" link whose two tasks are both on screen. */
export function dependencyArrows(rows: readonly TimelineRow[]): TimelineArrow[] {
  const rowOf = new Map<string, { row: number; bar: TimelineBar }>();
  rows.forEach((row, index) => {
    if (row.kind === 'task') rowOf.set(row.bar.task.id, { row: index, bar: row.bar });
  });
  const arrows: TimelineArrow[] = [];
  for (const [taskId, target] of rowOf) {
    for (const blockerId of target.bar.task.dependencies ?? []) {
      const source = rowOf.get(blockerId);
      if (!source) continue;
      arrows.push({
        fromTaskId: blockerId,
        toTaskId: taskId,
        fromRow: source.row,
        toRow: target.row,
        fromDay: source.bar.offset + source.bar.span,
        toDay: target.bar.offset,
      });
    }
  }
  return arrows;
}
