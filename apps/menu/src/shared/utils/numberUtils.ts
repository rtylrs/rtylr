/**
 * Generic number utility functions for the Menu app.
 */

/** Clamp a value between min and max (inclusive). */
export function clamp(value: number, min: number, max: number): number {
  return Math.min(Math.max(value, min), max);
}

/** Return true if the value is a finite number. */
export function isFiniteNumber(value: unknown): value is number {
  return typeof value === 'number' && Number.isFinite(value);
}

/** Return the percentage that part is of whole, or 0 when whole is 0. */
export function pct(part: number, whole: number): number {
  if (whole === 0) return 0;
  return (part / whole) * 100;
}

/** Round a number to a given number of decimal places. */
export function round(value: number, decimals = 0): number {
  const factor = Math.pow(10, decimals);
  return Math.round(value * factor) / factor;
}

/** Return the sum of an array of numbers. */
export function sum(values: number[]): number {
  return values.reduce((acc, v) => acc + v, 0);
}

/** Return the average of an array of numbers, or 0 for an empty array. */
export function average(values: number[]): number {
  if (values.length === 0) return 0;
  return sum(values) / values.length;
}

/** Return the minimum value in an array, or undefined when empty. */
export function min(values: number[]): number | undefined {
  if (values.length === 0) return undefined;
  return Math.min(...values);
}

/** Return the maximum value in an array, or undefined when empty. */
export function max(values: number[]): number | undefined {
  if (values.length === 0) return undefined;
  return Math.max(...values);
}
