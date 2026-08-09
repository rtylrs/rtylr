import { describe, it, expect } from 'vitest';

import {
  unique,
  uniqueBy,
  chunk,
  flatten,
  groupBy,
  difference,
  intersection,
  last,
  first,
  moveItem,
} from './arrayUtils';

describe('unique', () => {
  it('removes duplicates', () => {
    expect(unique([1, 2, 2, 3, 1])).toEqual([1, 2, 3]);
  });

  it('preserves order', () => {
    expect(unique(['b', 'a', 'b'])).toEqual(['b', 'a']);
  });
});

describe('uniqueBy', () => {
  it('deduplicates by key function', () => {
    const items = [{ id: 1 }, { id: 2 }, { id: 1 }];
    expect(uniqueBy(items, (i) => i.id).length).toBe(2);
  });
});

describe('chunk', () => {
  it('splits array into chunks', () => {
    expect(chunk([1, 2, 3, 4, 5], 2)).toEqual([[1, 2], [3, 4], [5]]);
  });

  it('returns empty for empty array', () => {
    expect(chunk([], 3)).toEqual([]);
  });
});

describe('flatten', () => {
  it('flattens one level', () => {
    expect(
      flatten([
        [1, 2],
        [3, 4],
      ])
    ).toEqual([1, 2, 3, 4]);
  });
});

describe('groupBy', () => {
  it('groups by key', () => {
    const items = [
      { type: 'a', val: 1 },
      { type: 'b', val: 2 },
      { type: 'a', val: 3 },
    ];
    const grouped = groupBy(items, (i) => i.type);
    expect(grouped.get('a')?.length).toBe(2);
    expect(grouped.get('b')?.length).toBe(1);
  });
});

describe('difference', () => {
  it('returns elements in first but not second', () => {
    expect(difference([1, 2, 3], [2, 3, 4])).toEqual([1]);
  });
});

describe('intersection', () => {
  it('returns elements present in both arrays', () => {
    expect(intersection([1, 2, 3], [2, 3, 4])).toEqual([2, 3]);
  });
});

describe('last', () => {
  it('returns last element', () => {
    expect(last([1, 2, 3])).toBe(3);
  });

  it('returns undefined for empty array', () => {
    expect(last([])).toBeUndefined();
  });
});

describe('first', () => {
  it('returns first element', () => {
    expect(first([1, 2, 3])).toBe(1);
  });

  it('returns undefined for empty array', () => {
    expect(first([])).toBeUndefined();
  });
});

describe('moveItem', () => {
  it('moves item from one index to another', () => {
    expect(moveItem([1, 2, 3, 4], 0, 2)).toEqual([2, 3, 1, 4]);
  });

  it('does not mutate the original array', () => {
    const arr = [1, 2, 3];
    moveItem(arr, 0, 2);
    expect(arr).toEqual([1, 2, 3]);
  });
});
