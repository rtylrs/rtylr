/**
 * Duration parsing and formatting for time logs.
 *
 * Lifted out of TaskDetail.tsx, where both functions lived inline. Every
 * surface that renders or accepts a duration reads from here, so `1:30` is
 * understood the same way on a task panel as it will be in a timesheet cell.
 */

const CLOCK = /^(\d+):([0-5]\d)$/;
const DECIMAL_HOURS = /^(\d+(?:\.\d+)?)h$/;
const MINUTES = /^(\d+)m?$/;

/**
 * Reads the three notations people actually type: `90` / `90m`, `1.5h`, and
 * `1:30`. Returns null for anything else - including zero, which is never an
 * intentional entry and would otherwise post a row that means nothing.
 */
export function parseDurationInput(raw: string): number | null {
  const trimmed = raw.trim().toLowerCase();

  const clock = trimmed.match(CLOCK);
  if (clock) {
    const minutes = parseInt(clock[1], 10) * 60 + parseInt(clock[2], 10);
    return minutes > 0 ? minutes : null;
  }

  const hours = trimmed.match(DECIMAL_HOURS);
  if (hours) {
    const minutes = Math.round(parseFloat(hours[1]) * 60);
    return minutes > 0 ? minutes : null;
  }

  const mins = trimmed.match(MINUTES);
  if (mins) {
    const minutes = parseInt(mins[1], 10);
    return minutes > 0 ? minutes : null;
  }

  return null;
}

/** Compact prose form: `45m`, `2h`, `1h 30m`. For inline copy. */
export function formatDuration(minutes: number): string {
  const h = Math.floor(minutes / 60);
  const m = minutes % 60;
  if (h === 0) return `${m}m`;
  if (m === 0) return `${h}h`;
  return `${h}h ${m}m`;
}

/**
 * Padded clock form: `6:45`, `0:05`. For figures that stack in a column, where
 * `1h 5m` and `12h 30m` would not align. Hours are never rolled over at 24 -
 * a monthly total legitimately exceeds it.
 */
export function formatDurationClock(minutes: number): string {
  const h = Math.floor(minutes / 60);
  const m = minutes % 60;
  return `${h}:${String(m).padStart(2, '0')}`;
}
