// Sparkline data preparation helpers for the insights analytics feature.

export interface SparklinePoint {
  value: number;
  label?: string;
}

/**
 * sparklineMin returns the minimum value in a sparkline series.
 */
export function sparklineMin(points: SparklinePoint[]): number {
  if (points.length === 0) return 0;
  return Math.min(...points.map((p) => p.value));
}

/**
 * sparklineMax returns the maximum value in a sparkline series.
 */
export function sparklineMax(points: SparklinePoint[]): number {
  if (points.length === 0) return 0;
  return Math.max(...points.map((p) => p.value));
}

/**
 * sparklineTrend returns 'up', 'down', or 'flat' based on first vs last value.
 */
export function sparklineTrend(points: SparklinePoint[]): 'up' | 'down' | 'flat' {
  if (points.length < 2) return 'flat';
  const first = points[0].value;
  const last = points[points.length - 1].value;
  if (last > first) return 'up';
  if (last < first) return 'down';
  return 'flat';
}

/**
 * sparklineValues extracts values from sparkline points.
 */
export function sparklineValues(points: SparklinePoint[]): number[] {
  return points.map((p) => p.value);
}

/**
 * normalizeSparkline returns points with values normalized to 0-100.
 */
export function normalizeSparkline(points: SparklinePoint[]): SparklinePoint[] {
  const min = sparklineMin(points);
  const max = sparklineMax(points);
  const range = max - min;
  if (range === 0) return points.map((p) => ({ ...p, value: 0 }));
  return points.map((p) => ({
    ...p,
    value: Math.round(((p.value - min) / range) * 100),
  }));
}

/**
 * lastValue returns the most recent value in the sparkline.
 */
export function lastValue(points: SparklinePoint[]): number {
  if (points.length === 0) return 0;
  return points[points.length - 1].value;
}
