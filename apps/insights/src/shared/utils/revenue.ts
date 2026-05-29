// Revenue series helpers for the insights revenue feature.

import type { RevenuePoint } from '../hooks/queries/types';

/**
 * totalRevenue returns the sum of all amountMinorUnits in the series.
 */
export function totalRevenue(series: RevenuePoint[]): number {
  return series.reduce((sum, p) => sum + p.amountMinorUnits, 0);
}

/**
 * avgDailyRevenue returns the average amountMinorUnits per data point.
 * Returns 0 for an empty series.
 */
export function avgDailyRevenue(series: RevenuePoint[]): number {
  if (series.length === 0) return 0;
  return Math.round(totalRevenue(series) / series.length);
}

/**
 * peakRevenueDay returns the RevenuePoint with the highest amountMinorUnits.
 * Returns undefined for an empty series.
 */
export function peakRevenueDay(series: RevenuePoint[]): RevenuePoint | undefined {
  if (series.length === 0) return undefined;
  return series.reduce((best, p) => (p.amountMinorUnits > best.amountMinorUnits ? p : best));
}

/**
 * revenueSeriesLabels extracts the date strings from a revenue series.
 */
export function revenueSeriesLabels(series: RevenuePoint[]): string[] {
  return series.map((p) => p.date);
}

/**
 * revenueSeriesValues extracts amountMinorUnits from a revenue series.
 */
export function revenueSeriesValues(series: RevenuePoint[]): number[] {
  return series.map((p) => p.amountMinorUnits);
}

/**
 * weeklyTotals groups the series into chunks of 7 and returns each chunk's total.
 * Remaining points less than 7 are included as the last group.
 */
export function weeklyTotals(series: RevenuePoint[]): number[] {
  const totals: number[] = [];
  for (let i = 0; i < series.length; i += 7) {
    const chunk = series.slice(i, i + 7);
    totals.push(chunk.reduce((s, p) => s + p.amountMinorUnits, 0));
  }
  return totals;
}
