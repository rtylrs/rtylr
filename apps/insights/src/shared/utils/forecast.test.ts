import { describe, it, expect } from 'vitest';
import {
  linearForecast,
  forecastAccuracy,
  growthForecast,
  forecastPeriodLabels,
  forecastValues,
} from './forecast';
import type { ForecastPoint } from './forecast';

function makePoint(period: string, predicted: number, actual?: number): ForecastPoint {
  return { period, predicted, actual };
}

describe('linearForecast', () => {
  it('projects next value using linear trend', () => {
    const next = linearForecast([10, 20, 30]);
    expect(next).toBe(40);
  });

  it('returns single value for single-element series', () => {
    expect(linearForecast([50])).toBe(50);
  });

  it('returns 0 for empty series', () => {
    expect(linearForecast([])).toBe(0);
  });
});

describe('forecastAccuracy', () => {
  it('returns high accuracy for accurate predictions', () => {
    const points = [makePoint('Jan', 100, 100), makePoint('Feb', 110, 110)];
    expect(forecastAccuracy(points)).toBe(100);
  });

  it('returns 0 for empty points', () => {
    expect(forecastAccuracy([])).toBe(0);
  });

  it('ignores points without actual values', () => {
    const points = [makePoint('Jan', 100), makePoint('Feb', 110, 100)];
    expect(forecastAccuracy(points)).toBeGreaterThan(0);
  });
});

describe('growthForecast', () => {
  it('applies compound growth rate', () => {
    expect(growthForecast(100, 10, 1)).toBe(110);
  });

  it('compounds over multiple periods', () => {
    expect(growthForecast(100, 10, 2)).toBeCloseTo(121, 0);
  });
});

describe('forecastPeriodLabels', () => {
  it('extracts period labels', () => {
    const points = [makePoint('Q1', 100), makePoint('Q2', 110)];
    expect(forecastPeriodLabels(points)).toEqual(['Q1', 'Q2']);
  });
});

describe('forecastValues', () => {
  it('extracts predicted values', () => {
    const points = [makePoint('Q1', 100), makePoint('Q2', 110)];
    expect(forecastValues(points)).toEqual([100, 110]);
  });
});
