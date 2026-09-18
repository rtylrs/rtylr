// Date timeframe selection helpers for the insights analytics feature.

export type Timeframe = '7d' | '30d' | '90d' | '6m' | '1y' | 'ytd' | 'custom';

export interface TimeframeRange {
  start: string;
  end: string;
}

/**
 * timeframeLabel returns a display-friendly label for a timeframe.
 */
export function timeframeLabel(timeframe: Timeframe | string): string {
  const labels: Record<string, string> = {
    '7d': 'Last 7 Days',
    '30d': 'Last 30 Days',
    '90d': 'Last 90 Days',
    '6m': 'Last 6 Months',
    '1y': 'Last Year',
    ytd: 'Year to Date',
    custom: 'Custom Range',
  };
  return labels[timeframe] ?? timeframe;
}

/**
 * allTimeframes returns all standard timeframe options.
 */
export function allTimeframes(): Timeframe[] {
  return ['7d', '30d', '90d', '6m', '1y', 'ytd', 'custom'];
}

/**
 * timeframeDays returns the approximate number of days for a timeframe.
 * Returns null for 'ytd' and 'custom' which require context.
 */
export function timeframeDays(timeframe: Timeframe): number | null {
  const days: Partial<Record<Timeframe, number>> = {
    '7d': 7,
    '30d': 30,
    '90d': 90,
    '6m': 180,
    '1y': 365,
  };
  return days[timeframe] ?? null;
}

/**
 * isCustomTimeframe returns true when the timeframe is custom.
 */
export function isCustomTimeframe(timeframe: Timeframe | string): boolean {
  return timeframe === 'custom';
}

/**
 * timeframeRangeValid returns true when the range has a start before the end.
 */
export function timeframeRangeValid(range: TimeframeRange): boolean {
  return new Date(range.start) < new Date(range.end);
}

/**
 * timeframeDayCount returns the number of days in a custom range.
 */
export function timeframeDayCount(range: TimeframeRange): number {
  const msPerDay = 1000 * 60 * 60 * 24;
  const diff = new Date(range.end).getTime() - new Date(range.start).getTime();
  return Math.round(diff / msPerDay);
}
