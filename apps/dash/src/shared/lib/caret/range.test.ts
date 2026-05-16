import { describe, expect, it } from 'vitest';
import { range } from './range';

describe('range', () => {
  it('produces an inclusive range', () => {
    expect(range(1, 4)).toEqual([1, 2, 3, 4]);
  });
  it('returns empty for step 0', () => {
    expect(range(1, 5, 0)).toEqual([]);
  });
  it('supports negative step', () => {
    expect(range(5, 1, -1)).toEqual([5, 4, 3, 2, 1]);
  });
});
