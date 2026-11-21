import { addDays, daysBetween, toLocalIso } from './dateUtils';
import type { MonthCell } from './monthGrid';

/**
 * Multi-day layout for the calendars: tasks that run from a start date to a due
 * date, cycles, and milestones, laid out as bars across a week row.
 *
 * Everything here works on `YYYY-MM-DD` strings, never on a parsed `Date`, for
 * the reason monthGrid.ts records: a bare date parsed as UTC midnight lands on
 * the day before west of Greenwich (rtylr#1192). `addDays` and `daysBetween`
 * are UTC in and UTC out, so they never touch the local offset.
 */

/** An inclusive run of calendar days. `start` is never after `end`. */
export interface DayRange {
  start: string;
  end: string;
}

const ISO_DAY = /^\d{4}-\d{2}-\d{2}$/;

/** The calendar day part of a stored date, or null when there is none. */
export function isoDay(value: string | null | undefined): string | null {
  if (!value) return null;
  const day = value.split('T')[0];
  return ISO_DAY.test(day) ? day : null;
}

/**
 * The days a task occupies on a calendar.
 *
 * Start and due make a bar. A due date alone is one day, as it always was. A
 * start date alone is one day too, so a task somebody started planning does
 * not vanish from the calendar. A start after the due date is a contradiction
 * the API does not prevent; the due date wins, because that is the date the
 * calendar has always shown.
 */
export function taskRange(task: {
  dueDate?: string | null;
  startDate?: string | null;
}): DayRange | null {
  const due = isoDay(task.dueDate);
  const start = isoDay(task.startDate);
  if (due && start) return start <= due ? { start, end: due } : { start: due, end: due };
  if (due) return { start: due, end: due };
  if (start) return { start, end: start };
  return null;
}

/** Neither a start nor a due date: the Unscheduled panel's tasks. */
export function isUnscheduled(task: {
  dueDate?: string | null;
  startDate?: string | null;
}): boolean {
  return taskRange(task) === null;
}

/** The seven Sunday-first days of the week holding `cursor`, all in range. */
export function weekCells(cursor: Date): MonthCell[] {
  const first = new Date(
    cursor.getFullYear(),
    cursor.getMonth(),
    cursor.getDate() - cursor.getDay()
  );
  return Array.from({ length: 7 }, (_, offset) => {
    const date = new Date(first.getFullYear(), first.getMonth(), first.getDate() + offset);
    return { date, iso: toLocalIso(date), inMonth: true };
  });
}

/** Splits a grid's cells into rows of seven. */
export function chunkWeeks<T>(cells: readonly T[]): T[][] {
  const weeks: T[][] = [];
  for (let i = 0; i < cells.length; i += 7) weeks.push(cells.slice(i, i + 7));
  return weeks;
}

/** One bar's piece inside one week row. */
export interface WeekSegment<T> {
  item: T;
  /** Zero-based column the piece starts in, counted from the week's first day. */
  col: number;
  /** How many columns it covers, at least one. */
  span: number;
  /** Zero-based row inside the week, so overlapping bars stack. */
  lane: number;
  /** The bar's real start falls in this week, so the start edge is shown. */
  startsHere: boolean;
  /** The bar's real end falls in this week, so the end edge is shown. */
  endsHere: boolean;
}

/**
 * Lays bars out across one week, stacking overlaps into lanes.
 *
 * Earlier starts go first and, among equal starts, longer bars, which is the
 * order that keeps long bars on low lanes and short ones packed underneath.
 * Each bar takes the lowest lane free for all of its columns.
 */
export function layoutWeek<T>(
  weekStart: string,
  items: readonly { item: T; range: DayRange }[]
): WeekSegment<T>[] {
  const weekEnd = addDays(weekStart, 6);
  const visible = items
    .filter(({ range }) => range.start <= weekEnd && range.end >= weekStart)
    .sort((a, b) => {
      if (a.range.start !== b.range.start) return a.range.start < b.range.start ? -1 : 1;
      return daysBetween(b.range.start, b.range.end) - daysBetween(a.range.start, a.range.end);
    });

  // lanes[n] holds the busy columns of lane n.
  const lanes: boolean[][] = [];
  const segments: WeekSegment<T>[] = [];
  for (const { item, range } of visible) {
    const from = range.start < weekStart ? weekStart : range.start;
    const to = range.end > weekEnd ? weekEnd : range.end;
    const col = daysBetween(weekStart, from);
    const span = daysBetween(from, to) + 1;
    let lane = lanes.findIndex((busy) => !busy.slice(col, col + span).some(Boolean));
    if (lane === -1) {
      lane = lanes.length;
      lanes.push(Array.from({ length: 7 }, () => false));
    }
    for (let c = col; c < col + span; c++) lanes[lane][c] = true;
    segments.push({
      item,
      col,
      span,
      lane,
      startsHere: range.start >= weekStart,
      endsHere: range.end <= weekEnd,
    });
  }
  return segments;
}

/** Lanes a week's layout needs. */
export function laneCount(segments: readonly WeekSegment<unknown>[]): number {
  return segments.reduce((max, s) => Math.max(max, s.lane + 1), 0);
}

/**
 * Bars hidden on each day of the week because they sit at or past `maxLanes`.
 * Index 0 is the week's first day.
 */
export function hiddenPerDay(
  segments: readonly WeekSegment<unknown>[],
  maxLanes: number
): number[] {
  const counts = Array.from({ length: 7 }, () => 0);
  for (const s of segments) {
    if (s.lane < maxLanes) continue;
    for (let c = s.col; c < s.col + s.span; c++) counts[c] += 1;
  }
  return counts;
}

/**
 * Drags one edge of a range to `day`.
 *
 * Dragging an edge past the other one stops at it rather than flipping the
 * bar inside out, so the result is always a valid range of at least one day.
 */
export function resizeRange(range: DayRange, edge: 'start' | 'end', day: string): DayRange {
  if (edge === 'start') return { start: day <= range.end ? day : range.end, end: range.end };
  return { start: range.start, end: day >= range.start ? day : range.start };
}

/** Moves a whole range so that it starts on `start`, keeping its length. */
export function moveRange(range: DayRange, start: string): DayRange {
  return { start, end: addDays(start, daysBetween(range.start, range.end)) };
}

/** Every day of a range clipped to `[from, to]`, in order. */
export function daysInRange(range: DayRange, from: string, to: string): string[] {
  const first = range.start < from ? from : range.start;
  const last = range.end > to ? to : range.end;
  const days: string[] = [];
  for (let day = first; day <= last; day = addDays(day, 1)) days.push(day);
  return days;
}

/** What a drag on the calendar is doing to a task. */
export type DragKind = 'move' | 'start' | 'end' | 'schedule';

/** The date fields a calendar drag writes. Absent means left as it is. */
export interface SchedulePatch {
  startDate?: string;
  dueDate?: string;
}

/**
 * The task update a drop on `day` makes, or null when it changes nothing.
 *
 * - `schedule` is a task from the Unscheduled panel: the day becomes its due
 *   date, the field every calendar already shows.
 * - `move` shifts the whole bar by however far the grabbed day travelled, so a
 *   two-day task stays two days long. Only the fields the task already has are
 *   written: a task with only a due date keeps only a due date.
 * - `start` and `end` drag one edge. Dragging the far edge of a one-day task
 *   turns it into a range, so the date it did not have is filled in from the
 *   bar it had.
 */
export function reschedule(
  task: { dueDate?: string | null; startDate?: string | null },
  kind: DragKind,
  day: string,
  grabDay?: string
): SchedulePatch | null {
  const range = taskRange(task);
  const hasDue = isoDay(task.dueDate) !== null;
  const hasStart = isoDay(task.startDate) !== null;

  if (kind === 'schedule' || range === null) {
    return isoDay(task.dueDate) === day ? null : { dueDate: day };
  }

  if (kind === 'move') {
    const delta = daysBetween(grabDay ?? range.start, day);
    if (delta === 0) return null;
    const moved = moveRange(range, addDays(range.start, delta));
    return {
      ...(hasStart ? { startDate: moved.start } : {}),
      ...(hasDue || !hasStart ? { dueDate: moved.end } : {}),
    };
  }

  const resized = resizeRange(range, kind, day);
  if (resized.start === range.start && resized.end === range.end) return null;
  // A start after the due date is drawn on the due date (see taskRange), so an
  // edge drag writes the start it drew rather than leave the contradiction.
  const startDrawnElsewhere = hasStart && isoDay(task.startDate) !== range.start;
  const patch: SchedulePatch = {};
  if (resized.start !== range.start || !hasStart || startDrawnElsewhere) {
    patch.startDate = resized.start;
  }
  if (resized.end !== range.end || !hasDue) patch.dueDate = resized.end;
  return patch;
}
