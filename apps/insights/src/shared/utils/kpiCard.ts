// KPI card display helpers for the insights analytics feature.

export type KpiChangeDirection = 'increase' | 'decrease' | 'neutral';

export interface KpiCardData {
  title: string;
  value: number;
  previousValue: number;
  unit?: string;
  prefix?: string;
}

/**
 * kpiChange returns the absolute change between current and previous values.
 */
export function kpiChange(data: KpiCardData): number {
  return data.value - data.previousValue;
}

/**
 * kpiChangePct returns the percentage change from previous to current value.
 */
export function kpiChangePct(data: KpiCardData): number {
  if (data.previousValue === 0) return 0;
  return Math.round(((data.value - data.previousValue) / data.previousValue) * 1000) / 10;
}

/**
 * kpiChangeDirection returns whether the KPI increased, decreased, or stayed neutral.
 */
export function kpiChangeDirection(data: KpiCardData): KpiChangeDirection {
  const change = kpiChange(data);
  if (change > 0) return 'increase';
  if (change < 0) return 'decrease';
  return 'neutral';
}

/**
 * formatKpiValue formats a KPI value with optional prefix and unit.
 */
export function formatKpiValue(data: KpiCardData): string {
  const prefix = data.prefix ?? '';
  const unit = data.unit ?? '';
  const formatted =
    data.value >= 1000
      ? `${Math.round(data.value / 1000)}K`
      : data.value.toLocaleString('en-US', { maximumFractionDigits: 2 });
  return `${prefix}${formatted}${unit}`;
}

/**
 * kpiChangeBadge returns a string representation of the change for display.
 */
export function kpiChangeBadge(data: KpiCardData): string {
  const pct = kpiChangePct(data);
  const sign = pct >= 0 ? '+' : '';
  return `${sign}${pct}%`;
}
