/**
 * Analytics computation utility helpers for the Dash app.
 * All functions are pure with no side-effects.
 */

/** Compute percentage of a part relative to a total (0-100). Returns 0 when total is 0. */
export function percentOf(part: number, total: number): number {
  if (total === 0) return 0;
  return Math.round((part / total) * 1000) / 10;
}

/** Compute the period-over-period change as a percentage. Returns null when previous is 0. */
export function periodChange(current: number, previous: number): number | null {
  if (previous === 0) return null;
  return Math.round(((current - previous) / Math.abs(previous)) * 1000) / 10;
}

/** Format a period-over-period change value as a string like "+12.5%" or "-3.0%". */
export function formatChange(change: number | null): string {
  if (change === null) return 'N/A';
  const sign = change >= 0 ? '+' : '';
  return `${sign}${change.toFixed(1)}%`;
}

/** Return true when the change represents an improvement (positive = better). */
export function isImprovement(change: number | null, higherIsBetter = true): boolean {
  if (change === null) return false;
  return higherIsBetter ? change > 0 : change < 0;
}

/** Compute average of a number array. Returns 0 for an empty array. */
export function average(values: number[]): number {
  if (values.length === 0) return 0;
  return values.reduce((s, v) => s + v, 0) / values.length;
}

/** Compute the median of a number array. Returns 0 for an empty array. */
export function median(values: number[]): number {
  if (values.length === 0) return 0;
  const sorted = [...values].sort((a, b) => a - b);
  const mid = Math.floor(sorted.length / 2);
  return sorted.length % 2 === 0 ? (sorted[mid - 1]! + sorted[mid]!) / 2 : sorted[mid]!;
}

/** Return the maximum value in a number array. Returns 0 for an empty array. */
export function maxValue(values: number[]): number {
  if (values.length === 0) return 0;
  return Math.max(...values);
}

/** Return the minimum value in a number array. Returns 0 for an empty array. */
export function minValue(values: number[]): number {
  if (values.length === 0) return 0;
  return Math.min(...values);
}

/** Sum an array of numbers. */
export function sum(values: number[]): number {
  return values.reduce((s, v) => s + v, 0);
}

/** Group numeric values by a string key and return a sum per key. */
export function groupSum<T extends Record<string, unknown>>(
  items: T[],
  keyField: keyof T,
  valueField: keyof T
): Record<string, number> {
  const result: Record<string, number> = {};
  for (const item of items) {
    const key = String(item[keyField]);
    const val = Number(item[valueField]) || 0;
    result[key] = (result[key] ?? 0) + val;
  }
  return result;
}
