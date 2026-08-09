import { describe, it, expect } from 'vitest';
import {
  movingAverageForecast,
  exponentialSmoothingForecast,
  averageDailyDemand,
  daysOfStockRemaining,
  safetyStockForDemand,
  reorderPoint,
} from './demandForecastUtils';
import type { SalesDataPoint } from './demandForecastUtils';

const makeSales = (quantities: number[]): SalesDataPoint[] =>
  quantities.map((q, i) => ({ date: `2026-05-${String(i + 1).padStart(2, '0')}`, quantity: q }));

describe('movingAverageForecast', () => {
  it('computes average of last N periods', () => {
    const sales = makeSales([10, 20, 30, 40, 50]);
    const forecast = movingAverageForecast(sales, 3);
    // Last 3: 30, 40, 50 -> avg = 40
    expect(forecast.forecastedDemand).toBe(40);
    expect(forecast.method).toBe('simple-moving-average');
  });

  it('returns 0 for empty sales', () => {
    expect(movingAverageForecast([]).forecastedDemand).toBe(0);
  });

  it('includes confidence interval', () => {
    const sales = makeSales([10, 20, 30]);
    const forecast = movingAverageForecast(sales);
    expect(forecast.confidenceInterval.low).toBeGreaterThanOrEqual(0);
    expect(forecast.confidenceInterval.high).toBeGreaterThan(forecast.forecastedDemand);
  });
});

describe('exponentialSmoothingForecast', () => {
  it('returns a forecast value', () => {
    const sales = makeSales([10, 15, 12, 18, 14]);
    const forecast = exponentialSmoothingForecast(sales);
    expect(forecast.forecastedDemand).toBeGreaterThan(0);
    expect(forecast.method).toBe('exponential-smoothing');
  });

  it('returns 0 for empty sales', () => {
    expect(exponentialSmoothingForecast([]).forecastedDemand).toBe(0);
  });

  it('includes 20% confidence interval', () => {
    const sales = makeSales([100]);
    const forecast = exponentialSmoothingForecast(sales);
    expect(forecast.forecastedDemand).toBe(100);
    expect(forecast.confidenceInterval.high).toBe(120);
    expect(forecast.confidenceInterval.low).toBe(80);
  });
});

describe('averageDailyDemand', () => {
  it('computes daily average', () => {
    expect(averageDailyDemand(300, 30)).toBe(10);
  });

  it('returns 0 when days is 0', () => {
    expect(averageDailyDemand(100, 0)).toBe(0);
  });
});

describe('daysOfStockRemaining', () => {
  it('estimates days remaining correctly', () => {
    expect(daysOfStockRemaining(100, 10)).toBe(10);
  });

  it('returns Infinity when daily demand is 0', () => {
    expect(daysOfStockRemaining(100, 0)).toBe(Infinity);
  });
});

describe('safetyStockForDemand', () => {
  it('computes safety stock with 95% service level', () => {
    const stock = safetyStockForDemand(10, 4, 0.95);
    // 1.645 * 10 * sqrt(4) = 1.645 * 20 = 32.9 -> ceil = 33
    expect(stock).toBe(33);
  });

  it('returns higher stock for higher service level', () => {
    const stock95 = safetyStockForDemand(10, 4, 0.95);
    const stock99 = safetyStockForDemand(10, 4, 0.99);
    expect(stock99).toBeGreaterThan(stock95);
  });
});

describe('reorderPoint', () => {
  it('computes reorder point from demand and lead time', () => {
    // 10 units/day * 7 days + 20 safety = 90
    expect(reorderPoint(10, 7, 20)).toBe(90);
  });

  it('adds safety stock to reorder point', () => {
    const withSafety = reorderPoint(10, 5, 50);
    const withoutSafety = reorderPoint(10, 5, 0);
    expect(withSafety).toBe(withoutSafety + 50);
  });
});
