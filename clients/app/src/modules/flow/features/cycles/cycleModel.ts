import type { ApiCycle } from '../../shared/services/flowApi';

export interface BurndownPoint {
  /** yyyy-mm-dd */
  date: string;
  /** Work left at the end of the day, or null for days still to come. */
  remaining: number | null;
  /** Where a straight line from the full scope to zero would be. */
  ideal: number;
}

export interface BurndownItem {
  /** How much the item weighs: 1 for a task count, its points otherwise. */
  weight: number;
  /** When it was finished, if it is. */
  completedAt?: string;
}

const DAY_MS = 24 * 60 * 60 * 1000;

function parseDay(value: string): number {
  return Date.UTC(
    Number(value.slice(0, 4)),
    Number(value.slice(5, 7)) - 1,
    Number(value.slice(8, 10))
  );
}

function formatDay(ms: number): string {
  return new Date(ms).toISOString().slice(0, 10);
}

/**
 * One point per day of the cycle. `remaining` counts an item as done from the
 * day its completedAt falls on (UTC), so a task finished at 23:00 counts that
 * day. Days after `today` have no actual value.
 */
export function buildBurndown(
  startDate: string,
  endDate: string,
  items: readonly BurndownItem[],
  today: string
): BurndownPoint[] {
  if (!startDate || !endDate) return [];
  const start = parseDay(startDate);
  const end = parseDay(endDate);
  if (Number.isNaN(start) || Number.isNaN(end) || end < start) return [];
  const todayMs = parseDay(today);
  const total = items.reduce((sum, item) => sum + item.weight, 0);
  const days = Math.round((end - start) / DAY_MS);
  const points: BurndownPoint[] = [];
  for (let i = 0; i <= days; i++) {
    const day = start + i * DAY_MS;
    const dayEnd = day + DAY_MS;
    const done = items.reduce((sum, item) => {
      if (!item.completedAt) return sum;
      const at = Date.parse(item.completedAt);
      return !Number.isNaN(at) && at < dayEnd ? sum + item.weight : sum;
    }, 0);
    points.push({
      date: formatDay(day),
      remaining: day <= todayMs ? total - done : null,
      ideal: days === 0 ? 0 : Math.round(total * (1 - i / days) * 10) / 10,
    });
  }
  return points;
}

export function cyclePercent(cycle: Pick<ApiCycle, 'scope' | 'completed'>): number {
  return cycle.scope > 0 ? Math.round((cycle.completed / cycle.scope) * 100) : 0;
}

/** The end date `weeks` after a start date, inclusive of the start day. */
export function endDateForLength(startDate: string, weeks: number): string {
  if (!startDate) return '';
  const start = parseDay(startDate);
  if (Number.isNaN(start)) return '';
  return formatDay(start + (weeks * 7 - 1) * DAY_MS);
}
