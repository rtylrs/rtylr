// Period utility for the insights app.
// Converts named period strings (e.g. "last_7_days", "this_month") into
// concrete ISO date ranges. Used by dashboard KPI cards and chart hooks.

export type PeriodId =
  | 'today'
  | 'yesterday'
  | 'last_7_days'
  | 'last_30_days'
  | 'last_90_days'
  | 'this_week'
  | 'this_month'
  | 'this_quarter'
  | 'this_year'
  | 'last_month'
  | 'last_year';

export interface DateRange {
  from: string;
  to: string;
}

function toIso(d: Date): string {
  return d.toISOString().slice(0, 10);
}

function utcDate(y: number, m: number, day: number): Date {
  return new Date(Date.UTC(y, m, day));
}

/**
 * Resolves a named period string to a concrete {from, to} date range.
 * All calculations are done in UTC to avoid timezone drift.
 */
export function resolvePeriod(period: PeriodId, now = new Date()): DateRange {
  const y = now.getUTCFullYear();
  const m = now.getUTCMonth();
  const d = now.getUTCDate();
  const today = utcDate(y, m, d);

  switch (period) {
    case 'today':
      return { from: toIso(today), to: toIso(today) };

    case 'yesterday': {
      const yest = utcDate(y, m, d - 1);
      return { from: toIso(yest), to: toIso(yest) };
    }

    case 'last_7_days':
      return { from: toIso(utcDate(y, m, d - 6)), to: toIso(today) };

    case 'last_30_days':
      return { from: toIso(utcDate(y, m, d - 29)), to: toIso(today) };

    case 'last_90_days':
      return { from: toIso(utcDate(y, m, d - 89)), to: toIso(today) };

    case 'this_week': {
      const dow = today.getUTCDay();
      const daysToMon = dow === 0 ? -6 : 1 - dow;
      return { from: toIso(utcDate(y, m, d + daysToMon)), to: toIso(today) };
    }

    case 'this_month':
      return { from: toIso(utcDate(y, m, 1)), to: toIso(today) };

    case 'this_quarter': {
      const qStart = Math.floor(m / 3) * 3;
      return { from: toIso(utcDate(y, qStart, 1)), to: toIso(today) };
    }

    case 'this_year':
      return { from: toIso(utcDate(y, 0, 1)), to: toIso(today) };

    case 'last_month': {
      const prevMonth = m === 0 ? 11 : m - 1;
      const prevYear = m === 0 ? y - 1 : y;
      const lastDay = new Date(Date.UTC(prevYear, prevMonth + 1, 0)).getUTCDate();
      return {
        from: toIso(utcDate(prevYear, prevMonth, 1)),
        to: toIso(utcDate(prevYear, prevMonth, lastDay)),
      };
    }

    case 'last_year':
      return { from: toIso(utcDate(y - 1, 0, 1)), to: toIso(utcDate(y - 1, 11, 31)) };

    default:
      return { from: toIso(utcDate(y, m, d - 29)), to: toIso(today) };
  }
}

/**
 * Returns a human-readable label for a period id.
 */
export function periodLabel(period: PeriodId): string {
  const labels: Record<PeriodId, string> = {
    today: 'Today',
    yesterday: 'Yesterday',
    last_7_days: 'Last 7 days',
    last_30_days: 'Last 30 days',
    last_90_days: 'Last 90 days',
    this_week: 'This week',
    this_month: 'This month',
    this_quarter: 'This quarter',
    this_year: 'This year',
    last_month: 'Last month',
    last_year: 'Last year',
  };
  return labels[period];
}

/**
 * Returns all available period ids in display order.
 */
export const ALL_PERIODS: PeriodId[] = [
  'today',
  'yesterday',
  'last_7_days',
  'last_30_days',
  'last_90_days',
  'this_week',
  'this_month',
  'this_quarter',
  'this_year',
  'last_month',
  'last_year',
];
