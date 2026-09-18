// Number formatting and rounding utilities for the insights app.
// Centralizes display logic so chart tooltips, KPI cards, and tables
// all format numbers identically.

/**
 * roundTo rounds a number to the specified decimal places.
 */
export function roundTo(value: number, decimals: number): number {
  const factor = Math.pow(10, decimals);
  return Math.round(value * factor) / factor;
}

/**
 * clamp constrains a value to the inclusive range [min, max].
 */
export function clamp(value: number, min: number, max: number): number {
  return Math.min(Math.max(value, min), max);
}

/**
 * formatCompact formats a number using K/M/B suffixes for display in
 * KPI cards and chart axis labels where space is limited.
 * e.g. 1500 -> "1.5K", 2_000_000 -> "2M"
 */
export function formatCompact(value: number): string {
  const abs = Math.abs(value);
  const sign = value < 0 ? '-' : '';
  if (abs >= 1_000_000_000) return `${sign}${roundTo(abs / 1_000_000_000, 1)}B`;
  if (abs >= 1_000_000) return `${sign}${roundTo(abs / 1_000_000, 1)}M`;
  if (abs >= 1_000) return `${sign}${roundTo(abs / 1_000, 1)}K`;
  return `${sign}${abs}`;
}

/**
 * safeDiv divides numerator by denominator and returns fallback when
 * the denominator is 0. Avoids NaN/Infinity in derived metrics.
 */
export function safeDiv(numerator: number, denominator: number, fallback = 0): number {
  if (denominator === 0) return fallback;
  return numerator / denominator;
}

/**
 * formatOrdinal returns the English ordinal suffix for an integer.
 * e.g. 1 -> "1st", 2 -> "2nd", 11 -> "11th", 21 -> "21st"
 */
export function formatOrdinal(n: number): string {
  const abs = Math.abs(Math.floor(n));
  const mod100 = abs % 100;
  const mod10 = abs % 10;
  let suffix: string;
  if (mod100 >= 11 && mod100 <= 13) {
    suffix = 'th';
  } else if (mod10 === 1) {
    suffix = 'st';
  } else if (mod10 === 2) {
    suffix = 'nd';
  } else if (mod10 === 3) {
    suffix = 'rd';
  } else {
    suffix = 'th';
  }
  return `${n}${suffix}`;
}

/**
 * lerp linearly interpolates between a and b by t in [0, 1].
 */
export function lerp(a: number, b: number, t: number): number {
  return a + (b - a) * t;
}
