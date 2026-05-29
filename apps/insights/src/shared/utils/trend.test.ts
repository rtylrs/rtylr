import { describe, it, expect } from 'vitest';
import { trendDirection, pctChange, rollingAverage, growthRate, trendLabel } from './trend';

describe('trendDirection', () => {
  it('returns up for positive change', () => {
    expect(trendDirection(10)).toBe('up');
  });

  it('returns down for negative change', () => {
    expect(trendDirection(-5)).toBe('down');
  });

  it('returns flat when within threshold', () => {
    expect(trendDirection(0.5)).toBe('flat');
    expect(trendDirection(-0.5)).toBe('flat');
  });

  it('respects custom threshold', () => {
    expect(trendDirection(3, 5)).toBe('flat');
    expect(trendDirection(6, 5)).toBe('up');
  });
});

describe('pctChange', () => {
  it('calculates positive change', () => {
    expect(pctChange(100, 120)).toBeCloseTo(20);
  });

  it('calculates negative change', () => {
    expect(pctChange(100, 80)).toBeCloseTo(-20);
  });

  it('returns 0 when previous is 0', () => {
    expect(pctChange(0, 50)).toBe(0);
  });
});

describe('rollingAverage', () => {
  it('returns correct rolling average', () => {
    const result = rollingAverage([1, 2, 3, 4, 5], 3);
    expect(result[2]).toBeCloseTo(2);
    expect(result[4]).toBeCloseTo(4);
  });

  it('returns empty for empty series', () => {
    expect(rollingAverage([], 3)).toEqual([]);
  });

  it('returns empty for invalid window', () => {
    expect(rollingAverage([1, 2, 3], 0)).toEqual([]);
  });
});

describe('growthRate', () => {
  it('returns period-over-period rates', () => {
    const rates = growthRate([100, 110, 121]);
    expect(rates[0]).toBeCloseTo(10);
    expect(rates[1]).toBeCloseTo(10);
  });

  it('returns empty for series with one item', () => {
    expect(growthRate([100])).toEqual([]);
  });
});

describe('trendLabel', () => {
  it('returns Stable for flat trend', () => {
    expect(trendLabel(0.5)).toBe('Stable');
  });

  it('shows positive sign for upward trend', () => {
    expect(trendLabel(15)).toContain('+15%');
  });

  it('shows negative for downward trend', () => {
    expect(trendLabel(-8)).toContain('-8%');
  });
});
