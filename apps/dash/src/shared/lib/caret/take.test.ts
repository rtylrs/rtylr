import { describe, expect, it } from 'vitest';
import { take, takeRight } from './take';

describe('take', () => {
  it('returns the first N', () => {
    expect(take([1, 2, 3, 4], 2)).toEqual([1, 2]);
  });
  it('clamps at length', () => {
    expect(take([1, 2], 5)).toEqual([1, 2]);
  });
  it('returns empty for N<=0', () => {
    expect(take([1, 2], 0)).toEqual([]);
  });
});

describe('takeRight', () => {
  it('returns the last N', () => {
    expect(takeRight([1, 2, 3, 4], 2)).toEqual([3, 4]);
  });
});
