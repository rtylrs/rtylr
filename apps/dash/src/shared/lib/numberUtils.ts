/**
 * Number formatting and arithmetic helpers for the Dash app.
 */

/** Round to a given number of decimal places. */
export function round(value: number, decimals = 2): number {
  const factor = Math.pow(10, decimals);
  return Math.round(value * factor) / factor;
}

/** Floor to a given number of decimal places. */
export function floorTo(value: number, decimals = 2): number {
  const factor = Math.pow(10, decimals);
  return Math.floor(value * factor) / factor;
}

/** Ceil to a given number of decimal places. */
export function ceilTo(value: number, decimals = 2): number {
  const factor = Math.pow(10, decimals);
  return Math.ceil(value * factor) / factor;
}

/** Linear interpolation between `a` and `b` by factor `t` (0-1). */
export function lerp(a: number, b: number, t: number): number {
  return a + (b - a) * t;
}

/** Map a value from [inMin, inMax] to [outMin, outMax]. */
export function mapRange(
  value: number,
  inMin: number,
  inMax: number,
  outMin: number,
  outMax: number
): number {
  if (inMax === inMin) return outMin;
  return ((value - inMin) / (inMax - inMin)) * (outMax - outMin) + outMin;
}

/** Compute the sum of an array of numbers. */
export function sum(values: number[]): number {
  return values.reduce((acc, v) => acc + v, 0);
}

/** Compute the arithmetic mean of an array. Returns 0 for empty arrays. */
export function mean(values: number[]): number {
  if (values.length === 0) return 0;
  return sum(values) / values.length;
}

/** Compute the median of an array. Returns 0 for empty arrays. */
export function median(values: number[]): number {
  if (values.length === 0) return 0;
  const sorted = [...values].sort((a, b) => a - b);
  const mid = Math.floor(sorted.length / 2);
  return sorted.length % 2 !== 0 ? sorted[mid] : (sorted[mid - 1] + sorted[mid]) / 2;
}

/** Check if a value is between `min` and `max` (inclusive). */
export function isBetween(value: number, min: number, max: number): boolean {
  return value >= min && value <= max;
}
