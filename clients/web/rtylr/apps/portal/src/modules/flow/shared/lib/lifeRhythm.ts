/**
 * Re-exported, not reimplemented.
 *
 * There were two functions called todayIso, this one local and dateUtils'
 * one UTC, so they returned different dates for hours every day and an
 * import from the wrong module was invisible at the call site. dateUtils has
 * the single definition now, and it is the local one this file had.
 */
import { todayIso } from './dateUtils';

export { todayIso };

export function isDueToday(dueDate: string | undefined): boolean {
  if (!dueDate) return false;
  return dueDate.slice(0, 10) === todayIso();
}

/** Monday 00:00 local time */
export function mondayStartMs(d = new Date()): number {
  const t = new Date(d);
  const day = t.getDay();
  const diff = t.getDate() - day + (day === 0 ? -6 : 1);
  t.setDate(diff);
  t.setHours(0, 0, 0, 0);
  return t.getTime();
}

/** Task due date falls in the current calendar week (Mon–Sun), date-only. */
export function isDueThisCalendarWeek(dueDate: string | undefined): boolean {
  if (!dueDate) return false;
  const start = mondayStartMs();
  const end = start + 7 * 86400000;
  const taskMs = new Date(dueDate + 'T12:00:00').getTime();
  return taskMs >= start && taskMs < end;
}

export function parseTimeToMinutes(hhmm: string): number {
  const [h, m] = hhmm.split(':').map((x) => parseInt(x, 10));
  if (Number.isNaN(h)) return 0;
  return h * 60 + (Number.isNaN(m) ? 0 : m);
}

/** Whether "now" falls inside quiet hours (supports overnight ranges e.g. 22:00–07:00). */
export function isQuietNow(quietStart: string, quietEnd: string): boolean {
  const now = new Date();
  const mins = now.getHours() * 60 + now.getMinutes();
  const a = parseTimeToMinutes(quietStart);
  const b = parseTimeToMinutes(quietEnd);
  if (a <= b) return mins >= a && mins < b;
  return mins >= a || mins < b;
}
