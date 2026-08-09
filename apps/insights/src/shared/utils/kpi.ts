// KPI card computation helpers for the insights dashboard.
// These functions derive the display values shown on summary cards
// from the raw API response shapes in shared/hooks/queries/types.ts.

import { toFiniteNumber } from '@voxire/api';

/**
 * deltaLabel formats a deltaPct value as a signed percentage string.
 * e.g. 12.3 -> "+12.3%", -5.0 -> "-5.0%", 0 -> "0.0%"
 */
export function deltaLabel(deltaPct: number, decimals = 1): string {
  const safeDelta = toFiniteNumber(deltaPct);
  const formatted = Math.abs(safeDelta).toFixed(decimals);
  if (safeDelta > 0) return `+${formatted}%`;
  if (safeDelta < 0) return `-${formatted}%`;
  return `0.${'0'.repeat(decimals)}%`;
}

/**
 * deltaDirection returns 'up', 'down', or 'flat' for a delta
 * percentage, using the threshold to avoid noise on tiny moves.
 */
export function deltaDirection(deltaPct: number, flatThreshold = 0.5): 'up' | 'down' | 'flat' {
  const safeDelta = toFiniteNumber(deltaPct);
  if (safeDelta > flatThreshold) return 'up';
  if (safeDelta < -flatThreshold) return 'down';
  return 'flat';
}

/**
 * kpiColor returns the semantic color name for a KPI delta, taking
 * whether higher-is-better into account.
 */
export function kpiColor(deltaPct: number, higherIsBetter = true): 'green' | 'red' | 'gray' {
  const dir = deltaDirection(deltaPct);
  if (dir === 'flat') return 'gray';
  if (dir === 'up') return higherIsBetter ? 'green' : 'red';
  return higherIsBetter ? 'red' : 'green';
}

/**
 * formatKpiValue formats a raw KPI value for display.
 * minorUnits: when true, divides by 100 and formats as money with currency.
 * Otherwise formats as a compact integer.
 */
export function formatKpiValue(
  value: number,
  options: { minorUnits?: boolean; currency?: string; locale?: string } = {}
): string {
  const { minorUnits = false, currency = 'USD', locale = 'en-US' } = options;
  if (minorUnits) {
    const major = toFiniteNumber(value) / 100;
    return major.toLocaleString(locale, {
      style: 'currency',
      currency,
      minimumFractionDigits: 0,
      maximumFractionDigits: 0,
    });
  }
  return toFiniteNumber(value).toLocaleString(locale);
}
