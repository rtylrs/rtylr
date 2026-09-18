import { describe, it, expect } from 'vitest';
import { previousPeriod, rangeLength, compareRanges, sameMonthLastYear } from './comparison';

describe('previousPeriod', () => {
  it('returns a range of the same length immediately before the input range', () => {
    const range = { from: '2026-02-01', to: '2026-02-28' };
    const prev = previousPeriod(range);
    expect(rangeLength(prev)).toBe(rangeLength(range));
    expect(prev.to).toBe('2026-01-31');
  });

  it('does not overlap with the original range', () => {
    const range = { from: '2026-03-01', to: '2026-03-31' };
    const prev = previousPeriod(range);
    expect(prev.to < range.from).toBe(true);
  });

  it('works for a single-day range', () => {
    const range = { from: '2026-05-15', to: '2026-05-15' };
    const prev = previousPeriod(range);
    expect(prev.from).toBe('2026-05-14');
    expect(prev.to).toBe('2026-05-14');
  });
});

describe('rangeLength', () => {
  it('returns the inclusive number of days', () => {
    expect(rangeLength({ from: '2026-01-01', to: '2026-01-31' })).toBe(31);
  });

  it('returns 1 for a single-day range', () => {
    expect(rangeLength({ from: '2026-05-15', to: '2026-05-15' })).toBe(1);
  });
});

describe('compareRanges', () => {
  it('calculates absolute and percentage difference', () => {
    const result = compareRanges(120, 100);
    expect(result.absolute).toBe(20);
    expect(result.pct).toBeCloseTo(20);
  });

  it('returns pct 0 when previous is 0', () => {
    const result = compareRanges(50, 0);
    expect(result.pct).toBe(0);
  });

  it('handles a decline', () => {
    const result = compareRanges(80, 100);
    expect(result.absolute).toBe(-20);
    expect(result.pct).toBeCloseTo(-20);
  });
});

describe('sameMonthLastYear', () => {
  it('shifts the year back by one', () => {
    const result = sameMonthLastYear({ from: '2026-05-01', to: '2026-05-31' });
    expect(result.from).toBe('2025-05-01');
    expect(result.to).toBe('2025-05-31');
  });

  it('preserves the month and day', () => {
    const result = sameMonthLastYear({ from: '2026-12-25', to: '2026-12-31' });
    expect(result.from).toBe('2025-12-25');
    expect(result.to).toBe('2025-12-31');
  });
});
