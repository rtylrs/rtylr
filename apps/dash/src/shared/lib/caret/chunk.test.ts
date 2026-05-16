import { describe, expect, it } from 'vitest';
import { chunk } from './chunk';

describe('chunk', () => {
  it('splits into evenly-sized chunks', () => {
    expect(chunk([1, 2, 3, 4, 5], 2)).toEqual([[1, 2], [3, 4], [5]]);
  });
  it('returns a single chunk for size 0', () => {
    expect(chunk([1, 2], 0)).toEqual([[1, 2]]);
  });
});
