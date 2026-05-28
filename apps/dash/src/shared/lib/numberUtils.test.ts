import { describe, it, expect } from 'vitest';
import {
  round,
  floorTo,
  ceilTo,
  lerp,
  mapRange,
  sum,
  mean,
  median,
  isBetween,
} from './numberUtils';

describe('round', () => {
  it('rounds to 2 decimal places by default', () => {
    expect(round(1.235)).toBe(1.24);
    expect(round(1.234)).toBe(1.23);
  });

  it('rounds to specified decimal places', () => {
    expect(round(1.5678, 3)).toBe(1.568);
  });

  it('handles zero', () => {
    expect(round(0)).toBe(0);
  });

  it('handles negative values', () => {
    expect(round(-1.235, 2)).toBe(-1.24);
  });
});

describe('floorTo', () => {
  it('floors to 2 decimal places', () => {
    expect(floorTo(1.999)).toBe(1.99);
  });

  it('handles whole numbers', () => {
    expect(floorTo(5.0)).toBe(5);
  });
});

describe('ceilTo', () => {
  it('ceils to 2 decimal places', () => {
    expect(ceilTo(1.001)).toBe(1.01);
  });

  it('handles whole numbers', () => {
    expect(ceilTo(5.0)).toBe(5);
  });
});

describe('lerp', () => {
  it('returns a at t=0', () => {
    expect(lerp(0, 100, 0)).toBe(0);
  });

  it('returns b at t=1', () => {
    expect(lerp(0, 100, 1)).toBe(100);
  });

  it('returns midpoint at t=0.5', () => {
    expect(lerp(0, 100, 0.5)).toBe(50);
  });

  it('works with negative ranges', () => {
    expect(lerp(-10, 10, 0.5)).toBe(0);
  });
});

describe('mapRange', () => {
  it('maps a value from one range to another', () => {
    expect(mapRange(5, 0, 10, 0, 100)).toBe(50);
  });

  it('maps minimum value correctly', () => {
    expect(mapRange(0, 0, 10, 0, 100)).toBe(0);
  });

  it('maps maximum value correctly', () => {
    expect(mapRange(10, 0, 10, 0, 100)).toBe(100);
  });

  it('returns outMin when inMin equals inMax', () => {
    expect(mapRange(5, 5, 5, 0, 100)).toBe(0);
  });
});

describe('sum', () => {
  it('sums an array of numbers', () => {
    expect(sum([1, 2, 3, 4, 5])).toBe(15);
  });

  it('returns 0 for empty array', () => {
    expect(sum([])).toBe(0);
  });

  it('handles negative numbers', () => {
    expect(sum([-1, 1])).toBe(0);
  });
});

describe('mean', () => {
  it('computes the arithmetic mean', () => {
    expect(mean([1, 2, 3, 4, 5])).toBe(3);
  });

  it('returns 0 for empty array', () => {
    expect(mean([])).toBe(0);
  });

  it('handles single element', () => {
    expect(mean([42])).toBe(42);
  });
});

describe('median', () => {
  it('finds the middle value for odd-length arrays', () => {
    expect(median([3, 1, 4, 1, 5])).toBe(3);
  });

  it('averages the two middle values for even-length arrays', () => {
    expect(median([1, 2, 3, 4])).toBe(2.5);
  });

  it('returns 0 for empty array', () => {
    expect(median([])).toBe(0);
  });

  it('handles single element', () => {
    expect(median([7])).toBe(7);
  });
});

describe('isBetween', () => {
  it('returns true for value within range', () => {
    expect(isBetween(5, 0, 10)).toBe(true);
  });

  it('returns true for value at min (inclusive)', () => {
    expect(isBetween(0, 0, 10)).toBe(true);
  });

  it('returns true for value at max (inclusive)', () => {
    expect(isBetween(10, 0, 10)).toBe(true);
  });

  it('returns false for value below min', () => {
    expect(isBetween(-1, 0, 10)).toBe(false);
  });

  it('returns false for value above max', () => {
    expect(isBetween(11, 0, 10)).toBe(false);
  });
});
