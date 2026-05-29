// Anomaly detection helpers for the insights analytics feature.

export interface DataPoint {
  date: string;
  value: number;
}

/**
 * mean returns the arithmetic mean of a number series.
 */
export function mean(values: number[]): number {
  if (values.length === 0) return 0;
  return values.reduce((sum, v) => sum + v, 0) / values.length;
}

/**
 * stdDev returns the population standard deviation of a number series.
 */
export function stdDev(values: number[]): number {
  if (values.length === 0) return 0;
  const avg = mean(values);
  const variance = values.reduce((sum, v) => sum + (v - avg) ** 2, 0) / values.length;
  return Math.sqrt(variance);
}

/**
 * zScore returns how many standard deviations a value is from the mean.
 */
export function zScore(value: number, avg: number, sd: number): number {
  if (sd === 0) return 0;
  return (value - avg) / sd;
}

/**
 * isAnomaly returns true when the absolute z-score exceeds the threshold.
 */
export function isAnomaly(value: number, avg: number, sd: number, threshold = 2): boolean {
  return Math.abs(zScore(value, avg, sd)) > threshold;
}

/**
 * detectAnomalies returns data points that are statistical anomalies.
 */
export function detectAnomalies(points: DataPoint[], threshold = 2): DataPoint[] {
  const values = points.map((p) => p.value);
  const avg = mean(values);
  const sd = stdDev(values);
  return points.filter((p) => isAnomaly(p.value, avg, sd, threshold));
}

/**
 * outlierBounds returns the lower and upper bounds for normal values.
 */
export function outlierBounds(values: number[], threshold = 2): { lower: number; upper: number } {
  const avg = mean(values);
  const sd = stdDev(values);
  return { lower: avg - threshold * sd, upper: avg + threshold * sd };
}
