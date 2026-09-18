import { describe, it, expect } from 'vitest';
import { mean, stdDev, zScore, isAnomaly, detectAnomalies, outlierBounds } from './anomaly';
import type { DataPoint } from './anomaly';

function makePoint(date: string, value: number): DataPoint {
  return { date, value };
}

describe('mean', () => {
  it('returns arithmetic mean', () => {
    expect(mean([2, 4, 6])).toBe(4);
  });

  it('returns 0 for empty array', () => {
    expect(mean([])).toBe(0);
  });
});

describe('stdDev', () => {
  it('calculates population standard deviation', () => {
    expect(stdDev([2, 4, 4, 4, 5, 5, 7, 9])).toBeCloseTo(2, 0);
  });

  it('returns 0 for empty array', () => {
    expect(stdDev([])).toBe(0);
  });

  it('returns 0 when all values are the same', () => {
    expect(stdDev([5, 5, 5])).toBe(0);
  });
});

describe('zScore', () => {
  it('returns 0 when std dev is 0', () => {
    expect(zScore(5, 5, 0)).toBe(0);
  });

  it('returns positive z-score for above-average value', () => {
    expect(zScore(10, 5, 2.5)).toBe(2);
  });
});

describe('isAnomaly', () => {
  it('returns true when z-score exceeds threshold', () => {
    expect(isAnomaly(10, 5, 2, 2)).toBe(true);
  });

  it('returns false when z-score is within threshold', () => {
    expect(isAnomaly(6, 5, 2, 2)).toBe(false);
  });
});

describe('detectAnomalies', () => {
  it('finds outlier data points', () => {
    const points = [
      makePoint('2026-01', 10),
      makePoint('2026-02', 11),
      makePoint('2026-03', 10),
      makePoint('2026-04', 9),
      makePoint('2026-05', 10),
      makePoint('2026-06', 11),
      makePoint('2026-07', 9),
      makePoint('2026-08', 10),
      makePoint('2026-09', 100),
    ];
    const anomalies = detectAnomalies(points);
    expect(anomalies.some((p) => p.date === '2026-09')).toBe(true);
  });

  it('returns empty for uniform data', () => {
    const uniform = [makePoint('A', 5), makePoint('B', 5), makePoint('C', 5)];
    expect(detectAnomalies(uniform)).toHaveLength(0);
  });
});

describe('outlierBounds', () => {
  it('returns lower and upper bounds', () => {
    const bounds = outlierBounds([10, 20, 30, 40, 50]);
    expect(bounds.lower).toBeLessThan(10);
    expect(bounds.upper).toBeGreaterThan(50);
  });
});
