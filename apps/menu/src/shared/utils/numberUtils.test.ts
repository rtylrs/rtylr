import { describe, it, expect } from 'vitest';
import { clamp, isFiniteNumber, pct, round, sum, average, min, max } from './numberUtils';

describe('clamp', () => {
  it('returns value when within range', () => {
    expect(clamp(5, 0, 10)).toBe(5);
  });

  it('returns min when value is below range', () => {
    expect(clamp(-5, 0, 10)).toBe(0);
  });

  it('returns max when value is above range', () => {
    expect(clamp(15, 0, 10)).toBe(10);
  });

  it('returns min when value equals min', () => {
    expect(clamp(0, 0, 10)).toBe(0);
  });

  it('returns max when value equals max', () => {
    expect(clamp(10, 0, 10)).toBe(10);
  });
});

describe('isFiniteNumber', () => {
  it('returns true for a regular number', () => {
    expect(isFiniteNumber(42)).toBe(true);
  });

  it('returns false for Infinity', () => {
    expect(isFiniteNumber(Infinity)).toBe(false);
  });

  it('returns false for NaN', () => {
    expect(isFiniteNumber(NaN)).toBe(false);
  });

  it('returns false for a string', () => {
    expect(isFiniteNumber('42')).toBe(false);
  });

  it('returns false for null', () => {
    expect(isFiniteNumber(null)).toBe(false);
  });
});

describe('pct', () => {
  it('calculates the percentage correctly', () => {
    expect(pct(25, 100)).toBe(25);
  });

  it('returns 0 when whole is 0', () => {
    expect(pct(5, 0)).toBe(0);
  });

  it('returns 50 for half', () => {
    expect(pct(1, 2)).toBe(50);
  });
});

describe('round', () => {
  it('rounds to the nearest integer by default', () => {
    expect(round(4.6)).toBe(5);
  });

  it('rounds to specified decimal places', () => {
    expect(round(3.14159, 2)).toBe(3.14);
  });

  it('rounds down when below midpoint', () => {
    expect(round(2.4)).toBe(2);
  });
});

describe('sum', () => {
  it('sums an array of numbers', () => {
    expect(sum([1, 2, 3, 4])).toBe(10);
  });

  it('returns 0 for an empty array', () => {
    expect(sum([])).toBe(0);
  });

  it('handles negative numbers', () => {
    expect(sum([5, -3, 2])).toBe(4);
  });
});

describe('average', () => {
  it('computes the average', () => {
    expect(average([1, 2, 3])).toBe(2);
  });

  it('returns 0 for an empty array', () => {
    expect(average([])).toBe(0);
  });
});

describe('min', () => {
  it('returns the minimum value', () => {
    expect(min([3, 1, 4, 1, 5])).toBe(1);
  });

  it('returns undefined for an empty array', () => {
    expect(min([])).toBeUndefined();
  });
});

describe('max', () => {
  it('returns the maximum value', () => {
    expect(max([3, 1, 4, 1, 5])).toBe(5);
  });

  it('returns undefined for an empty array', () => {
    expect(max([])).toBeUndefined();
  });
});
