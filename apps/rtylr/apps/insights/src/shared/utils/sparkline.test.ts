import { describe, it, expect } from 'vitest';
import {
  sparklineMin,
  sparklineMax,
  sparklineTrend,
  sparklineValues,
  normalizeSparkline,
  lastValue,
} from './sparkline';
import type { SparklinePoint } from './sparkline';

function makePoint(value: number, label?: string): SparklinePoint {
  return { value, label };
}

const points = [makePoint(10), makePoint(30), makePoint(20), makePoint(50)];

describe('sparklineMin', () => {
  it('returns minimum value', () => {
    expect(sparklineMin(points)).toBe(10);
  });

  it('returns 0 for empty array', () => {
    expect(sparklineMin([])).toBe(0);
  });
});

describe('sparklineMax', () => {
  it('returns maximum value', () => {
    expect(sparklineMax(points)).toBe(50);
  });

  it('returns 0 for empty array', () => {
    expect(sparklineMax([])).toBe(0);
  });
});

describe('sparklineTrend', () => {
  it('returns up when last is above first', () => {
    expect(sparklineTrend(points)).toBe('up');
  });

  it('returns down when last is below first', () => {
    expect(sparklineTrend([makePoint(50), makePoint(30)])).toBe('down');
  });

  it('returns flat for single point', () => {
    expect(sparklineTrend([makePoint(10)])).toBe('flat');
  });
});

describe('sparklineValues', () => {
  it('extracts values from points', () => {
    expect(sparklineValues(points)).toEqual([10, 30, 20, 50]);
  });
});

describe('normalizeSparkline', () => {
  it('normalizes values to 0-100', () => {
    const normalized = normalizeSparkline(points);
    expect(normalized[0].value).toBe(0);
    expect(normalized[3].value).toBe(100);
  });

  it('returns all zeros for uniform data', () => {
    const uniform = [makePoint(5), makePoint(5), makePoint(5)];
    expect(normalizeSparkline(uniform).every((p) => p.value === 0)).toBe(true);
  });
});

describe('lastValue', () => {
  it('returns the last value', () => {
    expect(lastValue(points)).toBe(50);
  });

  it('returns 0 for empty array', () => {
    expect(lastValue([])).toBe(0);
  });
});
