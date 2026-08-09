/**
 * Numeric helpers for the ERP application.
 * All functions are pure with no side-effects.
 */

/** Clamps a number to the inclusive range [min, max]. */
export function clamp(value: number, min: number, max: number): number {
  return Math.min(Math.max(value, min), max);
}

/** Rounds a number to the specified number of decimal places. */
export function roundTo(value: number, decimals: number): number {
  const factor = Math.pow(10, decimals);
  return Math.round(value * factor) / factor;
}

/** Returns true if the value is a finite number (not NaN, Infinity, or a non-number type). */
export function isFiniteNumber(value: unknown): value is number {
  return typeof value === 'number' && isFinite(value);
}

/**
 * Computes the percentage change from one value to another.
 * Returns 0 when both are 0, or 100 when from is 0 and to is positive.
 */
export function percentageChange(from: number, to: number): number {
  if (from === 0 && to === 0) return 0;
  if (from === 0) return 100;
  return roundTo(((to - from) / Math.abs(from)) * 100, 10);
}

/** Returns the average of an array of numbers. Returns 0 for empty arrays. */
export function average(values: number[]): number {
  if (values.length === 0) return 0;
  return values.reduce((acc, v) => acc + v, 0) / values.length;
}

/**
 * Computes the weighted average of values with corresponding weights.
 * Returns 0 if arrays are empty or lengths differ.
 */
export function weightedAverage(values: number[], weights: number[]): number {
  if (values.length === 0 || values.length !== weights.length) return 0;
  const totalWeight = weights.reduce((acc, w) => acc + w, 0);
  if (totalWeight === 0) return 0;
  const weightedSum = values.reduce((acc, v, i) => acc + v * weights[i], 0);
  return weightedSum / totalWeight;
}

/** Returns the sum of an array of numbers. Returns 0 for empty arrays. */
export function sum(values: number[]): number {
  return values.reduce((acc, v) => acc + v, 0);
}

/**
 * Parses a float from a string, returning a fallback value if parsing fails.
 * Defaults to 0 if no fallback is provided.
 */
export function safeParseFloat(value: string, fallback: number = 0): number {
  const parsed = parseFloat(value);
  return isNaN(parsed) ? fallback : parsed;
}

/** Returns 1 for positive, -1 for negative, and 0 for zero. */
export function sign(value: number): number {
  if (value > 0) return 1;
  if (value < 0) return -1;
  return 0;
}
