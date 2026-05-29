import { describe, it, expect } from 'vitest';
import {
  percentChange,
  trendDirection,
  isTrendPositive,
  formatTrendPct,
  normalizeSeries,
  movingAverage,
  cumulativeSum,
  seriesGrowthRate,
} from './metrics';

describe('percentChange', () => {
  it('returns the correct percentage for positive growth', () => {
    expect(percentChange(120, 100)).toBeCloseTo(20);
  });

  it('returns a negative percentage for decline', () => {
    expect(percentChange(80, 100)).toBeCloseTo(-20);
  });

  it('returns 0 when previous is 0', () => {
    expect(percentChange(50, 0)).toBe(0);
  });

  it('handles negative previous value', () => {
    expect(percentChange(-50, -100)).toBeCloseTo(50);
  });
});

describe('trendDirection', () => {
  it('returns up for positive change', () => {
    expect(trendDirection(110, 100)).toBe('up');
  });

  it('returns down for negative change', () => {
    expect(trendDirection(90, 100)).toBe('down');
  });

  it('returns flat when change is within threshold', () => {
    expect(trendDirection(100.2, 100)).toBe('flat');
  });

  it('uses custom flat threshold', () => {
    expect(trendDirection(103, 100, 5)).toBe('flat');
    expect(trendDirection(106, 100, 5)).toBe('up');
  });

  it('returns flat when previous is 0', () => {
    expect(trendDirection(0, 0)).toBe('flat');
  });
});

describe('isTrendPositive', () => {
  it('returns true for up when higher is better', () => {
    expect(isTrendPositive('up', true)).toBe(true);
  });

  it('returns false for down when higher is better', () => {
    expect(isTrendPositive('down', true)).toBe(false);
  });

  it('returns true for down when lower is better', () => {
    expect(isTrendPositive('down', false)).toBe(true);
  });

  it('returns null for flat direction', () => {
    expect(isTrendPositive('flat')).toBeNull();
  });
});

describe('formatTrendPct', () => {
  it('includes a + sign for positive change', () => {
    expect(formatTrendPct(120, 100)).toBe('+20.0%');
  });

  it('includes a minus sign for negative change', () => {
    expect(formatTrendPct(80, 100)).toBe('-20.0%');
  });

  it('returns N/A when previous is 0', () => {
    expect(formatTrendPct(50, 0)).toBe('N/A');
  });

  it('respects custom decimal places', () => {
    expect(formatTrendPct(110, 100, 2)).toBe('+10.00%');
  });
});

describe('normalizeSeries', () => {
  it('normalizes values to [0, 1]', () => {
    const result = normalizeSeries([0, 50, 100]);
    expect(result[0]).toBe(0);
    expect(result[1]).toBeCloseTo(0.5);
    expect(result[2]).toBe(1);
  });

  it('returns all zeros when min equals max', () => {
    expect(normalizeSeries([5, 5, 5])).toEqual([0, 0, 0]);
  });

  it('returns an empty array for empty input', () => {
    expect(normalizeSeries([])).toEqual([]);
  });
});

describe('movingAverage', () => {
  it('calculates a 3-period moving average', () => {
    const result = movingAverage([1, 2, 3, 4, 5], 3);
    expect(result[2]).toBeCloseTo(2);
    expect(result[4]).toBeCloseTo(4);
  });

  it('handles window larger than series length', () => {
    const result = movingAverage([10, 20], 5);
    expect(result[0]).toBe(10);
    expect(result[1]).toBe(15);
  });

  it('returns an empty array for empty input', () => {
    expect(movingAverage([], 3)).toEqual([]);
  });

  it('returns the same series for window of 1', () => {
    expect(movingAverage([1, 2, 3], 1)).toEqual([1, 2, 3]);
  });
});

describe('cumulativeSum', () => {
  it('computes running totals', () => {
    expect(cumulativeSum([1, 2, 3, 4])).toEqual([1, 3, 6, 10]);
  });

  it('handles negative values', () => {
    expect(cumulativeSum([5, -3, 2])).toEqual([5, 2, 4]);
  });

  it('returns an empty array for empty input', () => {
    expect(cumulativeSum([])).toEqual([]);
  });
});

describe('seriesGrowthRate', () => {
  it('calculates growth from first to last value', () => {
    expect(seriesGrowthRate([100, 110, 120, 130])).toBeCloseTo(30);
  });

  it('returns a negative rate for declining series', () => {
    expect(seriesGrowthRate([100, 90, 80])).toBeCloseTo(-20);
  });

  it('returns null for series shorter than 2 elements', () => {
    expect(seriesGrowthRate([100])).toBeNull();
    expect(seriesGrowthRate([])).toBeNull();
  });

  it('returns null when the first value is 0', () => {
    expect(seriesGrowthRate([0, 50, 100])).toBeNull();
  });
});
