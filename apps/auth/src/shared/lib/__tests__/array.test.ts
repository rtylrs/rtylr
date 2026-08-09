import { describe, it, expect } from 'vitest';
import { unique, chunk, last, first, groupBy, sameElements, difference } from '../array';

describe('unique', () => {
  it('removes duplicate primitives', () => {
    expect(unique([1, 2, 2, 3, 1])).toEqual([1, 2, 3]);
  });

  it('preserves order of first occurrences', () => {
    expect(unique(['b', 'a', 'b'])).toEqual(['b', 'a']);
  });

  it('returns empty array for empty input', () => {
    expect(unique([])).toEqual([]);
  });
});

describe('chunk', () => {
  it('splits into even chunks', () => {
    expect(chunk([1, 2, 3, 4], 2)).toEqual([
      [1, 2],
      [3, 4],
    ]);
  });

  it('handles remainder items in last chunk', () => {
    expect(chunk([1, 2, 3, 4, 5], 2)).toEqual([[1, 2], [3, 4], [5]]);
  });

  it('returns empty array for size <= 0', () => {
    expect(chunk([1, 2, 3], 0)).toEqual([]);
  });
});

describe('last', () => {
  it('returns the last element', () => {
    expect(last([1, 2, 3])).toBe(3);
  });

  it('returns undefined for empty array', () => {
    expect(last([])).toBeUndefined();
  });
});

describe('first', () => {
  it('returns the first element', () => {
    expect(first([10, 20, 30])).toBe(10);
  });

  it('returns undefined for empty array', () => {
    expect(first([])).toBeUndefined();
  });
});

describe('groupBy', () => {
  it('groups items by key function', () => {
    const items = [
      { type: 'a', val: 1 },
      { type: 'b', val: 2 },
      { type: 'a', val: 3 },
    ];
    const result = groupBy(items, (i) => i.type);
    expect(result['a']).toHaveLength(2);
    expect(result['b']).toHaveLength(1);
  });
});

describe('sameElements', () => {
  it('returns true for arrays with same elements in different order', () => {
    expect(sameElements([1, 2, 3], [3, 1, 2])).toBe(true);
  });

  it('returns false for arrays with different lengths', () => {
    expect(sameElements([1, 2], [1, 2, 3])).toBe(false);
  });

  it('returns false for arrays with different elements', () => {
    expect(sameElements([1, 2, 3], [1, 2, 4])).toBe(false);
  });
});

describe('difference', () => {
  it('returns elements in a that are not in b', () => {
    expect(difference([1, 2, 3, 4], [2, 4])).toEqual([1, 3]);
  });

  it('returns empty array when a is subset of b', () => {
    expect(difference([1, 2], [1, 2, 3])).toEqual([]);
  });
});
