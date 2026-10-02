import { describe, expect, it } from 'vitest';
import {
  clamp,
  formatCurrency,
  formatPercent,
  isInRange,
  maxOf,
  mean,
  minOf,
  percentOf,
  roundTo,
  sum,
} from './hrNumberUtils';

describe('clamp', () => {
  it('clamps value to range', () => {
    expect(clamp(5, 0, 10)).toBe(5);
    expect(clamp(-5, 0, 10)).toBe(0);
    expect(clamp(15, 0, 10)).toBe(10);
  });
});

describe('roundTo', () => {
  it('rounds to specified decimals', () => {
    expect(roundTo(3.14159, 2)).toBeCloseTo(3.14);
    expect(roundTo(10, 0)).toBe(10);
  });
});

describe('percentOf', () => {
  it('calculates percentage', () => {
    expect(percentOf(50, 200)).toBeCloseTo(25);
    expect(percentOf(0, 100)).toBe(0);
    expect(percentOf(100, 100)).toBe(100);
  });

  it('returns 0 when total is 0', () => {
    expect(percentOf(10, 0)).toBe(0);
  });
});

describe('sum', () => {
  it('sums array of numbers', () => {
    expect(sum([1, 2, 3, 4, 5])).toBe(15);
    expect(sum([])).toBe(0);
  });
});

describe('mean', () => {
  it('calculates mean of array', () => {
    expect(mean([1, 2, 3, 4, 5])).toBe(3);
    expect(mean([10, 20])).toBe(15);
  });

  it('returns 0 for empty array', () => {
    expect(mean([])).toBe(0);
  });
});

describe('minOf', () => {
  it('returns minimum value', () => {
    expect(minOf([3, 1, 4, 1, 5])).toBe(1);
  });

  it('returns 0 for empty array', () => {
    expect(minOf([])).toBe(0);
  });
});

describe('maxOf', () => {
  it('returns maximum value', () => {
    expect(maxOf([3, 1, 4, 1, 5])).toBe(5);
  });

  it('returns 0 for empty array', () => {
    expect(maxOf([])).toBe(0);
  });
});

describe('formatCurrency', () => {
  it('formats currency in USD', () => {
    const result = formatCurrency(1234.56);
    expect(result).toContain('1,234.56');
  });

  it('formats the configured tenant currency', () => {
    expect(formatCurrency(1234.56, 'EUR', 'en-US')).toContain('€1,234.56');
    expect(formatCurrency(1234.56, 'GBP', 'en-US')).toContain('£1,234.56');
  });
});

describe('formatPercent', () => {
  it('formats ratio as percentage', () => {
    expect(formatPercent(0.5)).toBe('50.0%');
    expect(formatPercent(1.0)).toBe('100.0%');
    expect(formatPercent(0.123, 2)).toBe('12.30%');
  });
});

describe('isInRange', () => {
  it('returns true when value is within range', () => {
    expect(isInRange(5, 0, 10)).toBe(true);
    expect(isInRange(0, 0, 10)).toBe(true);
    expect(isInRange(10, 0, 10)).toBe(true);
  });

  it('returns false when value is outside range', () => {
    expect(isInRange(-1, 0, 10)).toBe(false);
    expect(isInRange(11, 0, 10)).toBe(false);
  });
});
