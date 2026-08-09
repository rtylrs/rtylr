/** Clamps a number between min and max inclusive. */
export function clamp(value: number, min: number, max: number): number {
  return Math.min(Math.max(value, min), max);
}

/** Returns the percentage of value relative to total (0-100), clamped to 0-100. */
export function toPercent(value: number, total: number): number {
  if (total === 0) return 0;
  return clamp(Math.round((value / total) * 100), 0, 100);
}

/** Rounds a number to N decimal places. */
export function roundTo(value: number, decimals: number): number {
  const factor = 10 ** decimals;
  return Math.round(value * factor) / factor;
}

/** Returns the sum of an array of numbers. */
export function sum(values: number[]): number {
  return values.reduce((acc, v) => acc + v, 0);
}

/** Returns the average of an array of numbers, or 0 if empty. */
export function average(values: number[]): number {
  if (values.length === 0) return 0;
  return sum(values) / values.length;
}

/** Returns the min value, or null for empty array. */
export function minOf(values: number[]): number | null {
  if (values.length === 0) return null;
  return Math.min(...values);
}

/** Returns the max value, or null for empty array. */
export function maxOf(values: number[]): number | null {
  if (values.length === 0) return null;
  return Math.max(...values);
}

/** Formats a number with thousands separators, e.g. 1000 -> "1,000". */
export function formatNumber(value: number): string {
  return toFiniteNumber(value).toLocaleString('en-US');
}

/** Formats a number as a compact string: 1000 -> "1K", 1000000 -> "1M". */
export function formatCompact(value: number): string {
  const safeValue = toFiniteNumber(value);
  if (safeValue >= 1_000_000) return `${roundTo(safeValue / 1_000_000, 1)}M`;
  if (safeValue >= 1_000) return `${roundTo(safeValue / 1_000, 1)}K`;
  return String(safeValue);
}

/** Returns true if value is a finite number (not NaN or Infinity). */
export function isFiniteNumber(value: unknown): value is number {
  return typeof value === 'number' && isFinite(value);
}

/** Returns the percentage change from old to new value. */
export function percentChange(oldValue: number, newValue: number): number {
  if (oldValue === 0) return newValue === 0 ? 0 : 100;
  return roundTo(((newValue - oldValue) / Math.abs(oldValue)) * 100, 1);
}
import { toFiniteNumber } from '@voxire/api';
