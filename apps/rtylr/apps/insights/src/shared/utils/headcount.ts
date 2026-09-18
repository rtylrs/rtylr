// Headcount analytics helpers for the insights operations feature.

import type { HeadcountResponse } from '../hooks/queries/types';
import { toFiniteNumber } from '@voxire/api';

/**
 * headcountTrend returns 'up', 'down', or 'flat' based on momDeltaPct.
 */
export function headcountTrend(data: HeadcountResponse): 'up' | 'down' | 'flat' {
  if (data.momDeltaPct > 0) return 'up';
  if (data.momDeltaPct < 0) return 'down';
  return 'flat';
}

/**
 * headcountDeltaLabel returns a formatted string like '+5 this month' or '-2 this month'.
 */
export function headcountDeltaLabel(data: HeadcountResponse): string {
  const sign = data.momDelta >= 0 ? '+' : '';
  return `${sign}${data.momDelta} this month`;
}

/**
 * headcountGrowthRate returns the month-over-month growth rate formatted as a percentage string.
 */
export function headcountGrowthRate(data: HeadcountResponse): string {
  const delta = toFiniteNumber(data.momDeltaPct);
  const sign = delta >= 0 ? '+' : '';
  return `${sign}${delta.toFixed(1)}%`;
}

/**
 * previousHeadcount estimates the headcount from last month by subtracting the delta.
 */
export function previousHeadcount(data: HeadcountResponse): number {
  return Math.max(0, data.activeStaff - data.momDelta);
}
