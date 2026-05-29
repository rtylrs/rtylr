/**
 * Chart data preparation helpers for the Dash analytics views.
 */

export interface DataPoint {
  label: string;
  value: number;
}

export function normalizeDataPoints(points: DataPoint[]): DataPoint[] {
  return points
    .filter((p) => Number.isFinite(p.value))
    .sort((a, b) => a.label.localeCompare(b.label));
}

export function percentChange(current: number, previous: number): number | null {
  if (previous === 0) return null;
  return ((current - previous) / Math.abs(previous)) * 100;
}

export function formatPercentChange(change: number | null, digits = 1): string {
  if (change === null) return 'N/A';
  const sign = change >= 0 ? '+' : '';
  return `${sign}${change.toFixed(digits)}%`;
}

export function padSeries(values: number[], targetLength: number): number[] {
  if (values.length >= targetLength) return values;
  return [...Array<number>(targetLength - values.length).fill(0), ...values];
}

export function movingAverage(values: number[], period: number): number[] {
  if (period <= 0 || values.length < period) return [];
  const result: number[] = [];
  for (let i = period - 1; i < values.length; i++) {
    const slice = values.slice(i - period + 1, i + 1);
    result.push(slice.reduce((s, v) => s + v, 0) / period);
  }
  return result;
}

export function aggregateByLabel(points: DataPoint[]): DataPoint[] {
  const map = new Map<string, number>();
  for (const p of points) {
    map.set(p.label, (map.get(p.label) ?? 0) + p.value);
  }
  return Array.from(map.entries()).map(([label, value]) => ({ label, value }));
}
