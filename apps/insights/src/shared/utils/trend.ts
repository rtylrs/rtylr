// Trend analysis helpers for the insights feature.

export type TrendDirection = 'up' | 'down' | 'flat';

/**
 * trendDirection classifies a percentage change as up, down, or flat.
 * Values within the flatThreshold are considered flat.
 */
export function trendDirection(pctChange: number, flatThreshold = 1): TrendDirection {
  if (Math.abs(pctChange) <= flatThreshold) return 'flat';
  return pctChange > 0 ? 'up' : 'down';
}

/**
 * pctChange returns the percentage change between two values.
 * Returns 0 when the previous value is 0.
 */
export function pctChange(prev: number, curr: number): number {
  if (prev === 0) return 0;
  return ((curr - prev) / Math.abs(prev)) * 100;
}

/**
 * rollingAverage returns the rolling N-period average for a series.
 */
export function rollingAverage(series: number[], window: number): number[] {
  if (window <= 0 || series.length === 0) return [];
  return series.map((_, i) => {
    const start = Math.max(0, i - window + 1);
    const slice = series.slice(start, i + 1);
    return slice.reduce((s, v) => s + v, 0) / slice.length;
  });
}

/**
 * growthRate returns period-over-period growth rates for a series.
 */
export function growthRate(series: number[]): number[] {
  const rates: number[] = [];
  for (let i = 1; i < series.length; i++) {
    rates.push(pctChange(series[i - 1], series[i]));
  }
  return rates;
}

/**
 * trendLabel returns a human-readable label for a trend direction with magnitude.
 */
export function trendLabel(pct: number): string {
  const direction = trendDirection(pct);
  if (direction === 'flat') return 'Stable';
  const sign = pct > 0 ? '+' : '';
  return `${sign}${Math.round(pct * 10) / 10}%`;
}
