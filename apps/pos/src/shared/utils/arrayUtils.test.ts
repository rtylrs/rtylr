import { describe, it, expect } from 'vitest';
import { groupBy, sumBy, maxBy, minBy, chunk, moveItem, sortByValue } from './arrayUtils';

describe('groupBy', () => {
  it('groups items by the result of the key function', () => {
    const items = [
      { name: 'Burger', category: 'Mains' },
      { name: 'Fries', category: 'Sides' },
      { name: 'Steak', category: 'Mains' },
    ];
    const result = groupBy(items, (i) => i.category);
    expect(result['Mains']).toHaveLength(2);
    expect(result['Sides']).toHaveLength(1);
  });

  it('returns an empty object for an empty array', () => {
    expect(groupBy([], (i: string) => i)).toEqual({});
  });
});

describe('sumBy', () => {
  it('sums a numeric field', () => {
    const items = [{ qty: 3 }, { qty: 7 }, { qty: 2 }];
    expect(sumBy(items, (i) => i.qty)).toBe(12);
  });

  it('returns 0 for an empty array', () => {
    expect(sumBy([], (i: { qty: number }) => i.qty)).toBe(0);
  });

  it('ignores non-finite values', () => {
    const items = [{ qty: 5 }, { qty: NaN }];
    expect(sumBy(items, (i) => i.qty)).toBe(5);
  });
});

describe('maxBy', () => {
  it('returns the item with the highest value', () => {
    const items = [{ price: 5 }, { price: 15 }, { price: 10 }];
    expect(maxBy(items, (i) => i.price)).toEqual({ price: 15 });
  });

  it('returns undefined for an empty array', () => {
    expect(maxBy([], (i: { price: number }) => i.price)).toBeUndefined();
  });

  it('returns the only item in a single-element array', () => {
    expect(maxBy([{ price: 9 }], (i) => i.price)).toEqual({ price: 9 });
  });
});

describe('minBy', () => {
  it('returns the item with the lowest value', () => {
    const items = [{ price: 5 }, { price: 15 }, { price: 10 }];
    expect(minBy(items, (i) => i.price)).toEqual({ price: 5 });
  });

  it('returns undefined for an empty array', () => {
    expect(minBy([], (i: { price: number }) => i.price)).toBeUndefined();
  });
});

describe('chunk', () => {
  it('splits an array into chunks of the given size', () => {
    expect(chunk([1, 2, 3, 4, 5], 2)).toEqual([[1, 2], [3, 4], [5]]);
  });

  it('returns the array as a single chunk when size >= length', () => {
    expect(chunk([1, 2, 3], 10)).toEqual([[1, 2, 3]]);
  });

  it('returns an empty array for an invalid size', () => {
    expect(chunk([1, 2, 3], 0)).toEqual([]);
  });

  it('handles an empty input array', () => {
    expect(chunk([], 2)).toEqual([]);
  });
});

describe('moveItem', () => {
  it('moves an item from one index to another', () => {
    expect(moveItem(['a', 'b', 'c'], 0, 2)).toEqual(['b', 'c', 'a']);
  });

  it('returns the same array when source equals target', () => {
    expect(moveItem([1, 2, 3], 1, 1)).toEqual([1, 2, 3]);
  });

  it('does not mutate the original array', () => {
    const original = [1, 2, 3];
    moveItem(original, 0, 2);
    expect(original).toEqual([1, 2, 3]);
  });

  it('returns the original array for out-of-range indices', () => {
    expect(moveItem([1, 2, 3], -1, 2)).toEqual([1, 2, 3]);
    expect(moveItem([1, 2, 3], 0, 10)).toEqual([1, 2, 3]);
  });
});

describe('sortByValue', () => {
  it('sorts ascending by the value function', () => {
    const items = [{ p: 15 }, { p: 5 }, { p: 10 }];
    expect(sortByValue(items, (i) => i.p)).toEqual([{ p: 5 }, { p: 10 }, { p: 15 }]);
  });

  it('does not mutate the original array', () => {
    const items = [{ p: 3 }, { p: 1 }];
    sortByValue(items, (i) => i.p);
    expect(items[0].p).toBe(3);
  });

  it('handles an empty array', () => {
    expect(sortByValue([], (i: { p: number }) => i.p)).toEqual([]);
  });
});
