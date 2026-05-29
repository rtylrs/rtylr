// Metric calculation utilities for the insights app.
// These handle percentage change, trend direction, and
// sparkline normalization for KPI cards and charts.

export type TrendDirection = 'up' | 'down' | 'flat';

/**
 * Calculates the percentage change from a previous value to a current value.
 * Returns 0 when the previous value is 0 to avoid division by zero.
 */
export function percentChange(current: number, previous: number): number {
  if (previous === 0) return 0;
  return ((current - previous) / Math.abs(previous)) * 100;
}

/**
 * Returns the trend direction for a metric.
 * A change within the flat threshold (default 0.5%) is considered flat.
 */
export function trendDirection(
  current: number,
  previous: number,
  flatThreshold = 0.5
): TrendDirection {
  const pct = percentChange(current, previous);
  if (Math.abs(pct) <= flatThreshold) return 'flat';
  return pct > 0 ? 'up' : 'down';
}

/**
 * Returns true when an upward trend is positive for this metric.
 * For metrics like "churn rate" or "expenses", up is bad.
 */
export function isTrendPositive(direction: TrendDirection, higherIsBetter = true): boolean | null {
  if (direction === 'flat') return null;
  return higherIsBetter ? direction === 'up' : direction === 'down';
}

/**
 * Formats a percentage change for display, e.g. "+12.5%" or "-3.2%".
 * Returns "N/A" when the previous value was 0.
 */
export function formatTrendPct(current: number, previous: number, decimals = 1): string {
  if (previous === 0) return 'N/A';
  const pct = percentChange(current, previous);
  const sign = pct >= 0 ? '+' : '';
  return `${sign}${pct.toFixed(decimals)}%`;
}

/**
 * Normalizes a series of values to the range [0, 1].
 * Returns all zeros when the series has no range (min === max).
 */
export function normalizeSeries(values: number[]): number[] {
  if (values.length === 0) return [];
  const min = Math.min(...values);
  const max = Math.max(...values);
  if (min === max) return values.map(() => 0);
  return values.map((v) => (v - min) / (max - min));
}

/**
 * Calculates a simple moving average over a window of the given size.
 * Pads with the first available value when the window is larger than the index.
 */
export function movingAverage(values: number[], window: number): number[] {
  if (values.length === 0) return [];
  return values.map((_, i) => {
    const start = Math.max(0, i - window + 1);
    const slice = values.slice(start, i + 1);
    return slice.reduce((a, b) => a + b, 0) / slice.length;
  });
}

/**
 * Returns the cumulative sum of a series.
 */
export function cumulativeSum(values: number[]): number[] {
  let acc = 0;
  return values.map((v) => {
    acc += v;
    return acc;
  });
}

/**
 * Computes growth rate between the first and last values of a series.
 * Returns null for series shorter than 2 elements or when the first value is 0.
 */
export function seriesGrowthRate(values: number[]): number | null {
  if (values.length < 2) return null;
  const first = values[0];
  if (first === 0) return null;
  const last = values[values.length - 1];
  return ((last - first) / Math.abs(first)) * 100;
}
