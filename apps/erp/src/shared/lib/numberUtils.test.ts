import { describe, it, expect } from 'vitest';
import {
  clamp,
  roundTo,
  isFiniteNumber,
  percentageChange,
  average,
  weightedAverage,
  sum,
  safeParseFloat,
  sign,
} from './numberUtils';

describe('clamp', () => {
  it('returns value when within range', () => {
    expect(clamp(5, 1, 10)).toBe(5);
  });

  it('returns min when below range', () => {
    expect(clamp(-1, 0, 10)).toBe(0);
  });

  it('returns max when above range', () => {
    expect(clamp(15, 0, 10)).toBe(10);
  });
});

describe('roundTo', () => {
  it('rounds to specified decimals', () => {
    expect(roundTo(1.2345, 2)).toBe(1.23);
    expect(roundTo(1.2355, 2)).toBe(1.24);
  });

  it('rounds to 0 decimals', () => {
    expect(roundTo(1.6, 0)).toBe(2);
  });
});

describe('isFiniteNumber', () => {
  it('returns true for finite numbers', () => {
    expect(isFiniteNumber(5)).toBe(true);
    expect(isFiniteNumber(0)).toBe(true);
  });

  it('returns false for non-numbers', () => {
    expect(isFiniteNumber(NaN)).toBe(false);
    expect(isFiniteNumber(Infinity)).toBe(false);
    expect(isFiniteNumber('5')).toBe(false);
    expect(isFiniteNumber(null)).toBe(false);
  });
});

describe('percentageChange', () => {
  it('calculates positive change', () => {
    expect(percentageChange(100, 150)).toBe(50);
  });

  it('calculates negative change', () => {
    expect(percentageChange(100, 80)).toBe(-20);
  });

  it('returns 0 when both are 0', () => {
    expect(percentageChange(0, 0)).toBe(0);
  });

  it('returns 100 when from is 0 and to is positive', () => {
    expect(percentageChange(0, 5)).toBe(100);
  });
});

describe('average', () => {
  it('returns average of values', () => {
    expect(average([2, 4, 6])).toBe(4);
  });

  it('returns 0 for empty array', () => {
    expect(average([])).toBe(0);
  });
});

describe('weightedAverage', () => {
  it('computes weighted average', () => {
    expect(weightedAverage([10, 20], [1, 3])).toBe(17.5);
  });

  it('returns 0 for empty arrays', () => {
    expect(weightedAverage([], [])).toBe(0);
  });

  it('returns 0 when lengths differ', () => {
    expect(weightedAverage([1, 2], [1])).toBe(0);
  });
});

describe('sum', () => {
  it('returns sum of values', () => {
    expect(sum([1, 2, 3, 4])).toBe(10);
  });

  it('returns 0 for empty array', () => {
    expect(sum([])).toBe(0);
  });
});

describe('safeParseFloat', () => {
  it('parses valid number string', () => {
    expect(safeParseFloat('3.14')).toBeCloseTo(3.14);
  });

  it('returns fallback for invalid string', () => {
    expect(safeParseFloat('abc')).toBe(0);
    expect(safeParseFloat('abc', 99)).toBe(99);
  });
});

describe('sign', () => {
  it('returns 1 for positive numbers', () => {
    expect(sign(5)).toBe(1);
  });

  it('returns -1 for negative numbers', () => {
    expect(sign(-3)).toBe(-1);
  });

  it('returns 0 for zero', () => {
    expect(sign(0)).toBe(0);
  });
});
