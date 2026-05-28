import { describe, it, expect } from 'vitest';
import {
  clamp,
  inRange,
  range,
  pageCount,
  pageBounds,
  lerp,
  normalize,
} from './rangeUtils';

describe('clamp', () => {
  it('returns value when within range', () => {
    expect(clamp(5, 0, 10)).toBe(5);
  });

  it('clamps to min when below range', () => {
    expect(clamp(-5, 0, 10)).toBe(0);
  });

  it('clamps to max when above range', () => {
    expect(clamp(15, 0, 10)).toBe(10);
  });

  it('returns min for NaN', () => {
    expect(clamp(NaN, 0, 10)).toBe(0);
  });
});

describe('inRange', () => {
  it('returns true for a value within range', () => {
    expect(inRange(5, 0, 10)).toBe(true);
  });

  it('returns true for boundary values', () => {
    expect(inRange(0, 0, 10)).toBe(true);
    expect(inRange(10, 0, 10)).toBe(true);
  });

  it('returns false for out-of-range values', () => {
    expect(inRange(-1, 0, 10)).toBe(false);
    expect(inRange(11, 0, 10)).toBe(false);
  });

  it('returns false for NaN', () => {
    expect(inRange(NaN, 0, 10)).toBe(false);
  });
});

describe('range', () => {
  it('returns an array from start to end-1', () => {
    expect(range(0, 5)).toEqual([0, 1, 2, 3, 4]);
  });

  it('returns an empty array when end <= start', () => {
    expect(range(5, 3)).toEqual([]);
    expect(range(5, 5)).toEqual([]);
  });

  it('returns an empty array for non-finite inputs', () => {
    expect(range(NaN, 5)).toEqual([]);
  });
});

describe('pageCount', () => {
  it('calculates page count correctly', () => {
    expect(pageCount(100, 10)).toBe(10);
    expect(pageCount(101, 10)).toBe(11);
  });

  it('returns 0 for pageSize <= 0', () => {
    expect(pageCount(100, 0)).toBe(0);
  });

  it('returns 0 for non-finite total', () => {
    expect(pageCount(NaN, 10)).toBe(0);
  });
});

describe('pageBounds', () => {
  it('returns correct start and end for page 0', () => {
    expect(pageBounds(0, 10)).toEqual({ start: 0, end: 10 });
  });

  it('returns correct start and end for page 2', () => {
    expect(pageBounds(2, 10)).toEqual({ start: 20, end: 30 });
  });
});

describe('lerp', () => {
  it('interpolates correctly at t=0', () => {
    expect(lerp(0, 100, 0)).toBe(0);
  });

  it('interpolates correctly at t=1', () => {
    expect(lerp(0, 100, 1)).toBe(100);
  });

  it('interpolates correctly at t=0.5', () => {
    expect(lerp(0, 100, 0.5)).toBe(50);
  });

  it('returns a for non-finite inputs', () => {
    expect(lerp(5, 100, NaN)).toBe(5);
  });
});

describe('normalize', () => {
  it('normalizes a value to [0, 1]', () => {
    expect(normalize(5, 0, 10)).toBeCloseTo(0.5);
    expect(normalize(0, 0, 10)).toBe(0);
    expect(normalize(10, 0, 10)).toBe(1);
  });

  it('returns 0 when min equals max', () => {
    expect(normalize(5, 5, 5)).toBe(0);
  });

  it('returns 0 for NaN value', () => {
    expect(normalize(NaN, 0, 10)).toBe(0);
  });
});
