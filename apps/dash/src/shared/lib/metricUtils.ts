/**
 * Metric utility helpers for the Dash analytics views.
 * All functions are pure with no side-effects.
 */

export type TrendDirection = 'up' | 'down' | 'flat';

/** Determine trend direction from a percent change value. */
export function trendDirection(change: number): TrendDirection {
  if (change > 0) return 'up';
  if (change < 0) return 'down';
  return 'flat';
}

/** Clamp a metric value to the [min, max] range. */
export function clampMetric(value: number, min: number, max: number): number {
  if (!Number.isFinite(value)) return min;
  return Math.min(Math.max(value, min), max);
}

/**
 * Format a count as a short string with K/M suffixes.
 * e.g. 1500 -> "1.5K", 2_500_000 -> "2.5M"
 */
export function shortCount(n: number): string {
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1).replace(/\.0$/, '')}M`;
  if (n >= 1_000) return `${(n / 1_000).toFixed(1).replace(/\.0$/, '')}K`;
  return String(Math.round(n));
}

/** Compute a conversion rate as a percentage (0-100). Returns 0 when denominator is 0. */
export function conversionRate(numerator: number, denominator: number): number {
  if (denominator === 0) return 0;
  return Math.min(100, (numerator / denominator) * 100);
}

/** Return true when a metric meets or exceeds its goal. */
export function meetsGoal(value: number, goal: number, higherIsBetter = true): boolean {
  return higherIsBetter ? value >= goal : value <= goal;
}

/** Return progress toward goal as a 0-100 percentage. */
export function goalProgress(value: number, goal: number): number {
  if (goal === 0) return value === 0 ? 100 : 0;
  return Math.min(100, Math.max(0, (value / goal) * 100));
}
