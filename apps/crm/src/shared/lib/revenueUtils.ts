/**
 * Revenue calculation and formatting utilities for CRM analytics.
 */

import { toFiniteNumber } from '@voxire/api';

/** Parses a revenue string like "$1,234.56" to a number. */
export function parseRevenue(value: string): number {
  const cleaned = value.replace(/[$,\s]/g, '');
  const parsed = parseFloat(cleaned);
  return isNaN(parsed) ? 0 : parsed;
}

/** Formats a number as a compact revenue string. */
export function formatRevenue(value: number): string {
  const safeValue = toFiniteNumber(value);
  if (safeValue >= 1_000_000) return `$${(safeValue / 1_000_000).toFixed(1)}M`;
  if (safeValue >= 1_000) return `$${(safeValue / 1_000).toFixed(1)}K`;
  return `$${safeValue.toFixed(2)}`;
}

/** Returns the percentage change from previous to current. */
export function percentChange(previous: number, current: number): number {
  if (previous === 0) return current === 0 ? 0 : 100;
  return ((current - previous) / previous) * 100;
}

/** Returns total revenue from an array of revenue strings. */
export function sumRevenue(values: string[]): number {
  return values.reduce((sum, v) => sum + parseRevenue(v), 0);
}

/** Returns the average revenue per customer (ARPC). */
export function arpc(totalRevenue: number, customerCount: number): number {
  if (customerCount === 0) return 0;
  return totalRevenue / customerCount;
}

/** Returns a formatted change string like "+12%" or "-5%". */
export function formatChange(pct: number): string {
  const safePct = toFiniteNumber(pct);
  const sign = safePct >= 0 ? '+' : '';
  return `${sign}${safePct.toFixed(1)}%`;
}
