import type { ApiCycle } from '../../shared/services/flowApi';
import { addDays, toLocalIso } from '../../shared/lib/dateUtils';

const ISO_DAY = /^\d{4}-\d{2}-\d{2}$/;
// A cycle is at most a few weeks; this only stops a bad date from looping.
const MAX_DAYS = 366;

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

/**
 * One point per day of the cycle, in the viewer's time zone. `remaining`
 * counts an item as done from the local day its completedAt falls on. Days
 * after `today` have no actual value.
 */
export function buildBurndown(
  startDate: string,
  endDate: string,
  items: readonly BurndownItem[],
  today: string
): BurndownPoint[] {
  if (!ISO_DAY.test(startDate) || !ISO_DAY.test(endDate) || endDate < startDate) return [];
  const doneDays = items.map((item) => {
    if (!item.completedAt) return null;
    const at = new Date(item.completedAt);
    return Number.isNaN(at.getTime()) ? null : toLocalIso(at);
  });
  const total = items.reduce((sum, item) => sum + item.weight, 0);
  const days: string[] = [];
  for (let day = startDate; day <= endDate && days.length <= MAX_DAYS; day = addDays(day, 1)) {
    days.push(day);
  }
  const span = days.length - 1;
  return days.map((day, i) => {
    const done = items.reduce((sum, item, j) => {
      const doneDay = doneDays[j];
      return doneDay !== null && doneDay <= day ? sum + item.weight : sum;
    }, 0);
    return {
      date: day,
      remaining: day <= today ? total - done : null,
      ideal: span === 0 ? 0 : Math.round(total * (1 - i / span) * 10) / 10,
    };
  });
}

export function cyclePercent(cycle: Pick<ApiCycle, 'scope' | 'completed'>): number {
  return cycle.scope > 0 ? Math.round((cycle.completed / cycle.scope) * 100) : 0;
}

/** The end date `weeks` after a start date, inclusive of the start day. */
export function endDateForLength(startDate: string, weeks: number): string {
  return ISO_DAY.test(startDate) ? addDays(startDate, weeks * 7 - 1) : '';
}

export interface VelocityEntry {
  id: string;
  name: string;
  committed: number;
  completed: number;
}

/**
 * Committed and completed points of the project's last `limit` completed
 * cycles, oldest first, and the average completed.
 */
export function buildVelocity(
  cycles: readonly ApiCycle[],
  limit = 6
): { entries: VelocityEntry[]; average: number } {
  const entries = cycles
    .filter((c) => c.status === 'completed')
    .sort((a, b) => (a.endDate < b.endDate ? -1 : a.endDate > b.endDate ? 1 : 0))
    .slice(-limit)
    .map((c) => ({
      id: c.id,
      name: c.name,
      committed: c.committedPoints ?? 0,
      completed: c.completedPoints ?? 0,
    }));
  const average =
    entries.length > 0
      ? Math.round((entries.reduce((sum, e) => sum + e.completed, 0) / entries.length) * 10) / 10
      : 0;
  return { entries, average };
}
