import { describe, it, expect } from 'vitest';

import {
  clamp,
  toPercent,
  roundTo,
  sum,
  average,
  minOf,
  maxOf,
  formatNumber,
  formatCompact,
  isFiniteNumber,
  percentChange,
} from './numberUtils';

describe('clamp', () => {
  it('clamps below min', () => {
    expect(clamp(-5, 0, 100)).toBe(0);
  });

  it('clamps above max', () => {
    expect(clamp(200, 0, 100)).toBe(100);
  });

  it('returns value within range unchanged', () => {
    expect(clamp(50, 0, 100)).toBe(50);
  });
});

describe('toPercent', () => {
  it('calculates percentage', () => {
    expect(toPercent(3, 10)).toBe(30);
  });

  it('returns 0 for zero total', () => {
    expect(toPercent(5, 0)).toBe(0);
  });

  it('clamps to 100 max', () => {
    expect(toPercent(15, 10)).toBe(100);
  });
});

describe('roundTo', () => {
  it('rounds to given decimal places', () => {
    expect(roundTo(3.14159, 2)).toBe(3.14);
  });

  it('handles 0 decimal places', () => {
    expect(roundTo(3.7, 0)).toBe(4);
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

describe('average', () => {
  it('returns average', () => {
    expect(average([1, 2, 3])).toBeCloseTo(2);
  });

  it('returns 0 for empty array', () => {
    expect(average([])).toBe(0);
  });
});

describe('minOf', () => {
  it('returns minimum value', () => {
    expect(minOf([3, 1, 4, 1, 5])).toBe(1);
  });

  it('returns null for empty array', () => {
    expect(minOf([])).toBeNull();
  });
});

describe('maxOf', () => {
  it('returns maximum value', () => {
    expect(maxOf([3, 1, 4, 1, 5])).toBe(5);
  });

  it('returns null for empty array', () => {
    expect(maxOf([])).toBeNull();
  });
});

describe('formatNumber', () => {
  it('formats with thousands separators', () => {
    expect(formatNumber(1000)).toBe('1,000');
    expect(formatNumber(1000000)).toBe('1,000,000');
  });
});

describe('formatCompact', () => {
  it('formats thousands as K', () => {
    expect(formatCompact(1500)).toBe('1.5K');
  });

  it('formats millions as M', () => {
    expect(formatCompact(2500000)).toBe('2.5M');
  });

  it('returns small numbers as-is', () => {
    expect(formatCompact(99)).toBe('99');
  });
});

describe('isFiniteNumber', () => {
  it('returns true for finite numbers', () => {
    expect(isFiniteNumber(42)).toBe(true);
    expect(isFiniteNumber(0)).toBe(true);
  });

  it('returns false for NaN, Infinity, and non-numbers', () => {
    expect(isFiniteNumber(NaN)).toBe(false);
    expect(isFiniteNumber(Infinity)).toBe(false);
    expect(isFiniteNumber('42')).toBe(false);
  });
});

describe('percentChange', () => {
  it('returns percent change from old to new', () => {
    expect(percentChange(100, 120)).toBe(20);
    expect(percentChange(100, 80)).toBe(-20);
  });

  it('handles 0 old value', () => {
    expect(percentChange(0, 0)).toBe(0);
    expect(percentChange(0, 5)).toBe(100);
  });
});
