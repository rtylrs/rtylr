// Utilization analytics helpers for the insights operations feature.

import type { UtilizationResponse } from '../hooks/queries/types';

/**
 * utilizationPct returns the utilization percentage, clamped to 0-100.
 */
export function utilizationPct(data: UtilizationResponse): number {
  return Math.min(100, Math.max(0, data.utilizationPct));
}

/**
 * utilizationLabel returns a text classification based on the percentage.
 */
export function utilizationLabel(pct: number): 'low' | 'normal' | 'high' | 'overloaded' {
  if (pct < 50) return 'low';
  if (pct < 80) return 'normal';
  if (pct <= 100) return 'high';
  return 'overloaded';
}

/**
 * scheduledHoursLabel returns a formatted string for the scheduled hours.
 */
export function scheduledHoursLabel(data: UtilizationResponse): string {
  return `${data.scheduledHours}h scheduled`;
}

/**
 * availableCapacity returns the number of available hours not yet scheduled.
 */
export function availableCapacity(data: UtilizationResponse): number {
  const used = data.availableHours * (data.utilizationPct / 100);
  return Math.max(0, Math.round(data.availableHours - used));
}

/**
 * capacityWarning returns true when utilization is at or above 80%.
 */
export function capacityWarning(data: UtilizationResponse): boolean {
  return data.utilizationPct >= 80;
}
