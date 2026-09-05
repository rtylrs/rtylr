/**
 * ERP report generation and data aggregation helpers.
 * All functions are pure with no side-effects.
 */

export type ReportPeriod = 'daily' | 'weekly' | 'monthly' | 'quarterly' | 'yearly';

export interface ReportDataPoint {
  label: string;
  value: number;
  date?: string;
}

export interface ReportSummary {
  total: number;
  average: number;
  min: number;
  max: number;
  count: number;
  trend: 'up' | 'down' | 'flat';
}

/** Compute summary statistics for a series of data points. */
export function summarizeDataPoints(points: ReportDataPoint[]): ReportSummary {
  if (points.length === 0) {
    return { total: 0, average: 0, min: 0, max: 0, count: 0, trend: 'flat' };
  }

  const values = points.map((p) => p.value);
  const total = values.reduce((s, v) => s + v, 0);
  const average = total / values.length;
  const min = Math.min(...values);
  const max = Math.max(...values);

  let trend: 'up' | 'down' | 'flat' = 'flat';
  if (points.length >= 2) {
    const first = points[0].value;
    const last = points[points.length - 1].value;
    if (last > first * 1.01) trend = 'up';
    else if (last < first * 0.99) trend = 'down';
  }

  return { total, average, min, max, count: points.length, trend };
}

/** Compute period-over-period growth percentage. */
export function periodGrowthPct(current: number, previous: number): number {
  if (previous === 0) return current > 0 ? 100 : 0;
  return ((current - previous) / previous) * 100;
}

/** Group data points by a time period. */
export function groupByPeriod(
  points: ReportDataPoint[],
  period: ReportPeriod
): Map<string, ReportDataPoint[]> {
  const map = new Map<string, ReportDataPoint[]>();
  for (const point of points) {
    const key = getPeriodKey(point.date ?? '', period);
    const group = map.get(key) ?? [];
    map.set(key, [...group, point]);
  }
  return map;
}

/** Extract a period key from an ISO date string. */
function getPeriodKey(dateStr: string, period: ReportPeriod): string {
  if (!dateStr) return 'unknown';
  const d = new Date(dateStr);
  switch (period) {
    case 'daily':
      return dateStr.slice(0, 10);
    // UTC getters throughout, to match the daily and monthly branches, which
    // slice the string and so never see a local offset. With local getters a
    // 2026-01-01 row was grouped under 2025 west of Greenwich.
    case 'weekly': {
      const dayOfWeek = d.getUTCDay();
      const monday = new Date(d);
      monday.setUTCDate(d.getUTCDate() - ((dayOfWeek + 6) % 7));
      return monday.toISOString().slice(0, 10);
    }
    case 'monthly':
      return dateStr.slice(0, 7);
    case 'quarterly': {
      const quarter = Math.floor(d.getUTCMonth() / 3) + 1;
      return `${d.getUTCFullYear()}-Q${quarter}`;
    }
    case 'yearly':
      return String(d.getUTCFullYear());
  }
}

/** Aggregate data points by summing values for each period. */
export function aggregateByPeriod(
  points: ReportDataPoint[],
  period: ReportPeriod
): ReportDataPoint[] {
  const groups = groupByPeriod(points, period);
  return Array.from(groups.entries())
    .map(([label, pts]) => ({
      label,
      value: pts.reduce((s, p) => s + p.value, 0),
    }))
    .sort((a, b) => a.label.localeCompare(b.label));
}

/** Compute a rolling N-period average. */
export function rollingAverage(values: number[], n: number): number[] {
  return values.map((_, i) => {
    const window = values.slice(Math.max(0, i - n + 1), i + 1);
    return window.reduce((s, v) => s + v, 0) / window.length;
  });
}

/** Format a report value as a compact display string. */
export function formatReportValue(value: number, prefix = '', suffix = ''): string {
  let formatted: string;
  if (Math.abs(value) >= 1_000_000) {
    formatted = `${(value / 1_000_000).toFixed(1)}M`;
  } else if (Math.abs(value) >= 1_000) {
    formatted = `${(value / 1_000).toFixed(1)}K`;
  } else {
    formatted = value.toFixed(value % 1 === 0 ? 0 : 2);
  }
  return `${prefix}${formatted}${suffix}`;
}
