import { describe, it, expect } from 'vitest';
import { groupBy, uniqueBy, chunk, sumBy, maxBy, minBy, flatten } from './array';

describe('groupBy', () => {
  it('groups items by key', () => {
    const items = [
      { dept: 'eng', name: 'Alice' },
      { dept: 'eng', name: 'Bob' },
      { dept: 'design', name: 'Carol' },
    ];
    const map = groupBy(items, (i) => i.dept);
    expect(map.get('eng')).toHaveLength(2);
    expect(map.get('design')).toHaveLength(1);
  });

  it('returns empty map for empty input', () => {
    expect(groupBy([], (i: { id: number }) => i.id).size).toBe(0);
  });
});

describe('uniqueBy', () => {
  it('removes duplicates by key', () => {
    const items = [{ id: 1 }, { id: 2 }, { id: 1 }];
    expect(uniqueBy(items, (i) => i.id)).toHaveLength(2);
  });

  it('preserves the first occurrence', () => {
    const items = [
      { id: 1, v: 'a' },
      { id: 1, v: 'b' },
    ];
    const result = uniqueBy(items, (i) => i.id);
    expect(result[0].v).toBe('a');
  });

  it('returns an empty array for empty input', () => {
    expect(uniqueBy([], (i: { id: number }) => i.id)).toEqual([]);
  });
});

describe('chunk', () => {
  it('splits an array into chunks of size n', () => {
    expect(chunk([1, 2, 3, 4, 5], 2)).toEqual([[1, 2], [3, 4], [5]]);
  });

  it('returns the whole array in one chunk when n >= length', () => {
    expect(chunk([1, 2], 5)).toEqual([[1, 2]]);
  });

  it('returns empty array for n <= 0', () => {
    expect(chunk([1, 2], 0)).toEqual([]);
  });

  it('returns empty array for empty input', () => {
    expect(chunk([], 3)).toEqual([]);
  });
});

describe('sumBy', () => {
  it('sums the values returned by the accessor', () => {
    const items = [{ v: 1 }, { v: 2 }, { v: 3 }];
    expect(sumBy(items, (i) => i.v)).toBe(6);
  });

  it('returns 0 for empty input', () => {
    expect(sumBy([], (i: { v: number }) => i.v)).toBe(0);
  });
});

describe('maxBy', () => {
  it('returns the item with the highest value', () => {
    const items = [{ v: 1 }, { v: 5 }, { v: 3 }];
    expect(maxBy(items, (i) => i.v)).toEqual({ v: 5 });
  });

  it('returns undefined for empty input', () => {
    expect(maxBy([], (i: { v: number }) => i.v)).toBeUndefined();
  });
});

describe('minBy', () => {
  it('returns the item with the lowest value', () => {
    const items = [{ v: 4 }, { v: 1 }, { v: 3 }];
    expect(minBy(items, (i) => i.v)).toEqual({ v: 1 });
  });

  it('returns undefined for empty input', () => {
    expect(minBy([], (i: { v: number }) => i.v)).toBeUndefined();
  });
});

describe('flatten', () => {
  it('flattens one level of nested arrays', () => {
    expect(flatten([[1, 2], [3, 4], [5]])).toEqual([1, 2, 3, 4, 5]);
  });

  it('returns empty array for empty input', () => {
    expect(flatten([])).toEqual([]);
  });

  it('handles arrays with single elements', () => {
    expect(flatten([[1], [2], [3]])).toEqual([1, 2, 3]);
  });
});
