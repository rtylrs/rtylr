// App revenue share helpers for the insights revenue feature.

import type { AppRevenueShare } from '../hooks/queries/types';
import { toFiniteNumber } from '@voxire/api';

/**
 * totalAppRevenue sums amountMinorUnits across all app revenue shares.
 */
export function totalAppRevenue(shares: AppRevenueShare[]): number {
  return shares.reduce((sum, s) => sum + toFiniteNumber(s.amountMinorUnits), 0);
}

/**
 * topApps returns the n apps with the highest revenue.
 */
export function topApps(shares: AppRevenueShare[], n: number): AppRevenueShare[] {
  return [...shares].sort((a, b) => b.amountMinorUnits - a.amountMinorUnits).slice(0, n);
}

/**
 * appLabels extracts the app name from each revenue share entry.
 */
export function appLabels(shares: AppRevenueShare[]): string[] {
  return shares.map((s) => s.app);
}

/**
 * appValues extracts the amountMinorUnits from each revenue share entry.
 */
export function appValues(shares: AppRevenueShare[]): number[] {
  return shares.map((s) => s.amountMinorUnits);
}

/**
 * findAppShare returns the revenue share for a specific app name.
 * Returns undefined when the app is not found.
 */
export function findAppShare(
  shares: AppRevenueShare[],
  appName: string
): AppRevenueShare | undefined {
  return shares.find((s) => s.app === appName);
}

/**
 * appPctLabels returns the pct field formatted as a percentage string for each entry.
 */
export function appPctLabels(shares: AppRevenueShare[]): string[] {
  return shares.map((s) => `${toFiniteNumber(s.pct).toFixed(1)}%`);
}
