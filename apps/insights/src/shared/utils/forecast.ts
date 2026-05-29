// Revenue forecasting helpers for the insights analytics feature.

export interface ForecastPoint {
  period: string;
  actual?: number;
  predicted: number;
}

/**
 * linearForecast projects values using simple linear regression on a numeric series.
 * Returns the predicted value for the next period after the series.
 */
export function linearForecast(values: number[]): number {
  const n = values.length;
  if (n === 0) return 0;
  if (n === 1) return values[0];

  const xs = values.map((_, i) => i);
  const meanX = xs.reduce((s, x) => s + x, 0) / n;
  const meanY = values.reduce((s, v) => s + v, 0) / n;

  const slope =
    xs.reduce((s, x, i) => s + (x - meanX) * (values[i] - meanY), 0) /
    xs.reduce((s, x) => s + (x - meanX) ** 2, 0);

  const intercept = meanY - slope * meanX;
  return Math.round((slope * n + intercept) * 100) / 100;
}

/**
 * forecastAccuracy returns the mean absolute percentage error between actual and predicted values.
 */
export function forecastAccuracy(points: ForecastPoint[]): number {
  const withActual = points.filter((p) => p.actual !== undefined && p.actual !== 0);
  if (withActual.length === 0) return 0;
  const mape =
    withActual.reduce((sum, p) => {
      return sum + Math.abs((p.actual! - p.predicted) / p.actual!);
    }, 0) / withActual.length;
  return Math.round((1 - mape) * 1000) / 10;
}

/**
 * growthForecast returns the expected value after applying a constant growth rate.
 */
export function growthForecast(current: number, growthRate: number, periods: number): number {
  return Math.round(current * (1 + growthRate / 100) ** periods * 100) / 100;
}

/**
 * forecastPeriodLabels extracts period labels from a forecast series.
 */
export function forecastPeriodLabels(points: ForecastPoint[]): string[] {
  return points.map((p) => p.period);
}

/**
 * forecastValues extracts predicted values from a forecast series.
 */
export function forecastValues(points: ForecastPoint[]): number[] {
  return points.map((p) => p.predicted);
}
