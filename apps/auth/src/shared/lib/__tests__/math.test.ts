import { describe, it, expect } from 'vitest';
import { clamp, lerp, roundTo, inRange, percentage, sum, average } from '../math';

describe('clamp', () => {
  it('returns the value when within range', () => {
    expect(clamp(5, 0, 10)).toBe(5);
  });

  it('returns min when value is below range', () => {
    expect(clamp(-1, 0, 10)).toBe(0);
  });

  it('returns max when value is above range', () => {
    expect(clamp(15, 0, 10)).toBe(10);
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

  it('clamps t to 0-1 range', () => {
    expect(lerp(0, 100, 2)).toBe(100);
  });
});

describe('roundTo', () => {
  it('rounds to 2 decimal places', () => {
    expect(roundTo(1.234, 2)).toBe(1.23);
  });

  it('rounds to 0 decimal places', () => {
    expect(roundTo(3.6, 0)).toBe(4);
  });
});

describe('inRange', () => {
  it('returns true for value at min boundary', () => {
    expect(inRange(0, 0, 10)).toBe(true);
  });

  it('returns true for value at max boundary', () => {
    expect(inRange(10, 0, 10)).toBe(true);
  });

  it('returns false for value below range', () => {
    expect(inRange(-1, 0, 10)).toBe(false);
  });

  it('returns false for value above range', () => {
    expect(inRange(11, 0, 10)).toBe(false);
  });
});

describe('percentage', () => {
  it('calculates percentage correctly', () => {
    expect(percentage(25, 100)).toBe(25);
  });

  it('returns 0 when total is 0', () => {
    expect(percentage(5, 0)).toBe(0);
  });

  it('can exceed 100 for values greater than total', () => {
    expect(percentage(150, 100)).toBe(150);
  });
});

describe('sum', () => {
  it('sums an array of numbers', () => {
    expect(sum([1, 2, 3, 4])).toBe(10);
  });

  it('returns 0 for an empty array', () => {
    expect(sum([])).toBe(0);
  });
});

describe('average', () => {
  it('returns the average', () => {
    expect(average([2, 4, 6])).toBe(4);
  });

  it('returns 0 for an empty array', () => {
    expect(average([])).toBe(0);
  });
});
