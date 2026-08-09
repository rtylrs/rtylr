import { describe, it, expect } from 'vitest';
import {
  groupBy,
  unique,
  uniqueBy,
  chunk,
  sumBy,
  maxBy,
  minBy,
  sortBy,
  findFirst,
} from './arrayUtils';

describe('groupBy', () => {
  const items = [
    { id: 1, type: 'a', value: 10 },
    { id: 2, type: 'b', value: 20 },
    { id: 3, type: 'a', value: 30 },
  ];

  it('groups items by key', () => {
    const grouped = groupBy(items, 'type');
    expect(grouped['a']).toHaveLength(2);
    expect(grouped['b']).toHaveLength(1);
  });
});

describe('unique', () => {
  it('removes duplicates', () => {
    expect(unique([1, 2, 2, 3, 3, 3])).toEqual([1, 2, 3]);
  });

  it('returns empty array for empty input', () => {
    expect(unique([])).toEqual([]);
  });
});

describe('uniqueBy', () => {
  const items = [
    { id: 1, name: 'a' },
    { id: 2, name: 'b' },
    { id: 1, name: 'c' },
  ];

  it('returns items with unique keys', () => {
    expect(uniqueBy(items, 'id')).toHaveLength(2);
  });

  it('keeps first occurrence', () => {
    const result = uniqueBy(items, 'id');
    expect(result[0].name).toBe('a');
  });
});

describe('chunk', () => {
  it('splits array into chunks', () => {
    const result = chunk([1, 2, 3, 4, 5], 2);
    expect(result).toEqual([[1, 2], [3, 4], [5]]);
  });

  it('returns empty for size <= 0', () => {
    expect(chunk([1, 2, 3], 0)).toEqual([]);
  });

  it('handles chunk larger than array', () => {
    expect(chunk([1, 2], 10)).toEqual([[1, 2]]);
  });
});

describe('sumBy', () => {
  const items = [{ v: 1 }, { v: 2 }, { v: 3 }];

  it('sums values by key', () => {
    expect(sumBy(items, 'v')).toBe(6);
  });

  it('returns 0 for empty array', () => {
    expect(sumBy([], 'v' as never)).toBe(0);
  });
});

describe('maxBy', () => {
  const items = [{ v: 5 }, { v: 10 }, { v: 3 }];

  it('returns maximum value', () => {
    expect(maxBy(items, 'v')).toBe(10);
  });

  it('returns undefined for empty array', () => {
    expect(maxBy([], 'v' as never)).toBeUndefined();
  });
});

describe('minBy', () => {
  const items = [{ v: 5 }, { v: 10 }, { v: 3 }];

  it('returns minimum value', () => {
    expect(minBy(items, 'v')).toBe(3);
  });

  it('returns undefined for empty array', () => {
    expect(minBy([], 'v' as never)).toBeUndefined();
  });
});

describe('sortBy', () => {
  const items = [
    { name: 'Zebra', n: 3 },
    { name: 'Apple', n: 1 },
    { name: 'Mango', n: 2 },
  ];

  it('sorts ascending by default', () => {
    const sorted = sortBy(items, 'name');
    expect(sorted[0].name).toBe('Apple');
  });

  it('sorts descending when specified', () => {
    const sorted = sortBy(items, 'n', 'desc');
    expect(sorted[0].n).toBe(3);
  });

  it('does not mutate original array', () => {
    sortBy(items, 'name');
    expect(items[0].name).toBe('Zebra');
  });
});

describe('findFirst', () => {
  const items = [{ id: 1 }, { id: 2 }, { id: 3 }];

  it('returns the first matching item', () => {
    expect(findFirst(items, (i) => i.id === 2)?.id).toBe(2);
  });

  it('returns undefined when no match', () => {
    expect(findFirst(items, (i) => i.id === 99)).toBeUndefined();
  });
});
