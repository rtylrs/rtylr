/**
 * Demand forecasting helpers for the ERP app.
 * All functions are pure with no side-effects.
 */

export interface SalesDataPoint {
  date: string;
  quantity: number;
}

export interface ForecastResult {
  forecastedDemand: number;
  confidenceInterval: { low: number; high: number };
  method: string;
}

/** Compute simple moving average demand from past sales. */
export function movingAverageForecast(sales: SalesDataPoint[], periods = 7): ForecastResult {
  if (sales.length === 0) {
    return {
      forecastedDemand: 0,
      confidenceInterval: { low: 0, high: 0 },
      method: 'simple-moving-average',
    };
  }

  const recent = sales.slice(-periods);
  const total = recent.reduce((s, p) => s + p.quantity, 0);
  const avg = total / recent.length;

  const variance = recent.reduce((s, p) => s + Math.pow(p.quantity - avg, 2), 0) / recent.length;
  const stdDev = Math.sqrt(variance);

  return {
    forecastedDemand: Math.round(avg),
    confidenceInterval: {
      low: Math.max(0, Math.round(avg - stdDev)),
      high: Math.round(avg + stdDev),
    },
    method: 'simple-moving-average',
  };
}

/** Compute exponential smoothing forecast (alpha = smoothing factor). */
export function exponentialSmoothingForecast(sales: SalesDataPoint[], alpha = 0.3): ForecastResult {
  if (sales.length === 0) {
    return {
      forecastedDemand: 0,
      confidenceInterval: { low: 0, high: 0 },
      method: 'exponential-smoothing',
    };
  }

  let smoothed = sales[0].quantity;
  for (let i = 1; i < sales.length; i++) {
    smoothed = alpha * sales[i].quantity + (1 - alpha) * smoothed;
  }

  const forecast = Math.round(smoothed);
  const margin = Math.round(forecast * 0.2);

  return {
    forecastedDemand: forecast,
    confidenceInterval: {
      low: Math.max(0, forecast - margin),
      high: forecast + margin,
    },
    method: 'exponential-smoothing',
  };
}

/** Compute average daily demand from total sales and days. */
export function averageDailyDemand(totalQuantity: number, days: number): number {
  if (days === 0) return 0;
  return totalQuantity / days;
}

/** Estimate days of stock remaining given current stock and daily demand. */
export function daysOfStockRemaining(currentStock: number, dailyDemand: number): number {
  if (dailyDemand <= 0) return Infinity;
  return Math.floor(currentStock / dailyDemand);
}

/** Compute recommended safety stock (standard deviation * Z-score * lead time). */
export function safetyStockForDemand(
  demandStdDev: number,
  leadTimeDays: number,
  serviceLevel = 0.95
): number {
  // Z-score approximation for common service levels
  const zScores: Record<number, number> = {
    0.9: 1.28,
    0.95: 1.645,
    0.99: 2.326,
  };
  const z = zScores[serviceLevel] ?? 1.645;
  return Math.ceil(z * demandStdDev * Math.sqrt(leadTimeDays));
}

/** Compute reorder point based on demand and lead time. */
export function reorderPoint(
  averageDailyDemand: number,
  leadTimeDays: number,
  safetyStock = 0
): number {
  return Math.ceil(averageDailyDemand * leadTimeDays + safetyStock);
}
