import { toLocalIso } from './dateUtils';

/**
 * Shared month-cell construction for the two calendars.
 *
 * Both built their own grid, and the two disagreed. The global calendar keyed
 * cells with `toLocalIso` and bucketed tasks on `dueDate.split('T')[0]`, so both
 * sides were the same calendar day the reader is standing in. The project
 * calendar keyed cells from a local Date but bucketed on
 * `new Date(dueDate).toDateString()`, which parses a bare `YYYY-MM-DD` as **UTC**
 * midnight and then reads it back in local time. West of Greenwich that lands on
 * the day before, so every task on that screen sat on the wrong day. Measured:
 * a task due 2026-09-08 bucketed as `Mon Sep 07` in New York, Los Angeles and
 * Honolulu, while the cell for it was keyed `Tue Sep 08`.
 *
 * It read as correct from Beirut, which is the same reason the suite was green
 * (rtylr#1192).
 *
 * Everything here is local-in, local-out. Nothing routes through
 * `toISOString()`.
 */
export interface MonthCell {
  /** Local midnight for the day this cell shows. */
  date: Date;
  /** `YYYY-MM-DD` in the reader's own timezone, and the bucket key. */
  iso: string;
  /** False for the leading and trailing days that pad the grid to whole weeks. */
  inMonth: boolean;
}

/**
 * The cells of a month grid, padded to whole Sunday-first weeks.
 *
 * `padOutside` controls what the padding is: the global calendar shows the
 * neighbouring days dimmed, the project calendar leaves them blank. Passing
 * `false` still returns the cells, so the caller keeps a fixed 7-column grid,
 * and marks them `inMonth: false`.
 */
export function monthCells(cursor: Date): MonthCell[] {
  const year = cursor.getFullYear();
  const month = cursor.getMonth();
  const startDay = new Date(year, month, 1).getDay();
  const daysInMonth = new Date(year, month + 1, 0).getDate();

  const cell = (date: Date, inMonth: boolean): MonthCell => ({
    date,
    iso: toLocalIso(date),
    inMonth,
  });

  const cells: MonthCell[] = [];
  for (let i = startDay - 1; i >= 0; i--) {
    cells.push(cell(new Date(year, month, -i), false));
  }
  for (let day = 1; day <= daysInMonth; day++) {
    cells.push(cell(new Date(year, month, day), true));
  }
  while (cells.length % 7 !== 0) {
    const last = cells[cells.length - 1].date;
    cells.push(cell(new Date(last.getFullYear(), last.getMonth(), last.getDate() + 1), false));
  }
  return cells;
}

/**
 * Tasks grouped by the local calendar day of their due date.
 *
 * The key is the date part of the stored value, taken as a string. A due date is
 * a calendar day, not an instant, so it is never parsed into a `Date` first:
 * doing that is what put the project calendar a day out west of Greenwich.
 */
export function tasksByLocalDay<T extends { dueDate?: string | null }>(
  tasks: readonly T[]
): Map<string, T[]> {
  const map = new Map<string, T[]>();
  for (const task of tasks) {
    if (!task.dueDate) continue;
    const key = task.dueDate.split('T')[0];
    const bucket = map.get(key);
    if (bucket) {
      bucket.push(task);
    } else {
      map.set(key, [task]);
    }
  }
  return map;
}

/**
 * The days of this month that actually have tasks, in order, for the agenda.
 *
 * An agenda lists days with something on them rather than every day of the
 * month; a phone showing 30 headings, most of them empty, is the month grid's
 * problem in a taller shape.
 */
export function agendaDays<T extends { dueDate?: string | null }>(
  cells: readonly MonthCell[],
  byDay: Map<string, T[]>
): { cell: MonthCell; tasks: T[] }[] {
  return cells
    .filter((c) => c.inMonth)
    .map((cell) => ({ cell, tasks: byDay.get(cell.iso) ?? [] }))
    .filter((entry) => entry.tasks.length > 0);
}
