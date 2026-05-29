import { describe, it, expect } from 'vitest';
import {
  normalizeDataPoints,
  percentChange,
  formatPercentChange,
  padSeries,
  movingAverage,
  aggregateByLabel,
} from './chartHelpers';

describe('normalizeDataPoints', () => {
  it('removes non-finite values', () => {
    const pts = [
      { label: 'a', value: 1 },
      { label: 'b', value: Infinity },
    ];
    expect(normalizeDataPoints(pts)).toHaveLength(1);
  });
  it('sorts by label', () => {
    const pts = [
      { label: 'c', value: 3 },
      { label: 'a', value: 1 },
    ];
    expect(normalizeDataPoints(pts)[0]!.label).toBe('a');
  });
});

describe('percentChange', () => {
  it('returns null when previous is 0', () => expect(percentChange(10, 0)).toBeNull());
  it('computes positive change', () => expect(percentChange(110, 100)).toBeCloseTo(10));
  it('computes negative change', () => expect(percentChange(90, 100)).toBeCloseTo(-10));
});

describe('formatPercentChange', () => {
  it('formats positive with plus sign', () => expect(formatPercentChange(10)).toBe('+10.0%'));
  it('formats negative', () => expect(formatPercentChange(-5)).toBe('-5.0%'));
  it('returns N/A for null', () => expect(formatPercentChange(null)).toBe('N/A'));
});

describe('padSeries', () => {
  it('pads with leading zeros', () => expect(padSeries([3, 4, 5], 5)).toEqual([0, 0, 3, 4, 5]));
  it('returns original when at length', () => expect(padSeries([1, 2, 3], 3)).toEqual([1, 2, 3]));
});

describe('movingAverage', () => {
  it('computes 3-period average', () => {
    const result = movingAverage([1, 2, 3, 4, 5], 3);
    expect(result[0]).toBeCloseTo(2);
    expect(result[1]).toBeCloseTo(3);
  });
  it('returns [] when period > length', () => expect(movingAverage([1, 2], 5)).toEqual([]));
});

describe('aggregateByLabel', () => {
  it('sums duplicate labels', () => {
    const pts = [
      { label: 'a', value: 10 },
      { label: 'a', value: 15 },
    ];
    const result = aggregateByLabel(pts);
    expect(result.find((p) => p.label === 'a')!.value).toBe(25);
  });
  it('returns [] for empty input', () => expect(aggregateByLabel([])).toEqual([]));
});
