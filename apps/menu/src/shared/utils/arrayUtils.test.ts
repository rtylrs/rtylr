import { describe, it, expect } from 'vitest';
import {
  unique,
  chunk,
  take,
  partition,
  flatten,
  isEmpty,
  replaceAt,
  removeAt,
} from './arrayUtils';

describe('unique', () => {
  it('removes duplicate primitives', () => {
    expect(unique([1, 2, 2, 3])).toEqual([1, 2, 3]);
  });

  it('returns the same array when all elements are unique', () => {
    expect(unique(['a', 'b', 'c'])).toEqual(['a', 'b', 'c']);
  });

  it('returns an empty array for empty input', () => {
    expect(unique([])).toEqual([]);
  });
});

describe('chunk', () => {
  it('splits an array into chunks of the given size', () => {
    expect(chunk([1, 2, 3, 4], 2)).toEqual([
      [1, 2],
      [3, 4],
    ]);
  });

  it('handles arrays that are not evenly divisible', () => {
    expect(chunk([1, 2, 3], 2)).toEqual([[1, 2], [3]]);
  });

  it('returns an empty array for size less than 1', () => {
    expect(chunk([1, 2], 0)).toEqual([]);
  });

  it('returns the whole array in one chunk when size exceeds length', () => {
    expect(chunk([1, 2], 10)).toEqual([[1, 2]]);
  });
});

describe('take', () => {
  it('returns the first N items', () => {
    expect(take([1, 2, 3, 4], 2)).toEqual([1, 2]);
  });

  it('returns the full array when N exceeds length', () => {
    expect(take([1, 2], 10)).toEqual([1, 2]);
  });

  it('returns an empty array for N of 0', () => {
    expect(take([1, 2], 0)).toEqual([]);
  });
});

describe('partition', () => {
  it('splits items into matching and non-matching', () => {
    const [evens, odds] = partition([1, 2, 3, 4], (n) => n % 2 === 0);
    expect(evens).toEqual([2, 4]);
    expect(odds).toEqual([1, 3]);
  });

  it('returns an empty array for the non-matching side when all pass', () => {
    const [pass, fail] = partition([2, 4], (n) => n % 2 === 0);
    expect(pass).toEqual([2, 4]);
    expect(fail).toEqual([]);
  });
});

describe('flatten', () => {
  it('flattens one level of nesting', () => {
    expect(
      flatten([
        [1, 2],
        [3, 4],
      ])
    ).toEqual([1, 2, 3, 4]);
  });

  it('returns an empty array for empty input', () => {
    expect(flatten([])).toEqual([]);
  });
});

describe('isEmpty', () => {
  it('returns true for an empty array', () => {
    expect(isEmpty([])).toBe(true);
  });

  it('returns false for a non-empty array', () => {
    expect(isEmpty([1])).toBe(false);
  });
});

describe('replaceAt', () => {
  it('replaces the element at the given index', () => {
    expect(replaceAt([1, 2, 3], 1, 99)).toEqual([1, 99, 3]);
  });

  it('does not mutate the original array', () => {
    const arr = [1, 2, 3];
    replaceAt(arr, 0, 0);
    expect(arr).toEqual([1, 2, 3]);
  });

  it('returns the original array unchanged for an out-of-bounds index', () => {
    expect(replaceAt([1, 2], 5, 99)).toEqual([1, 2]);
  });
});

describe('removeAt', () => {
  it('removes the element at the given index', () => {
    expect(removeAt([1, 2, 3], 1)).toEqual([1, 3]);
  });

  it('does not mutate the original array', () => {
    const arr = [1, 2, 3];
    removeAt(arr, 0);
    expect(arr).toEqual([1, 2, 3]);
  });

  it('returns the original array for an out-of-bounds index', () => {
    expect(removeAt([1, 2], 5)).toEqual([1, 2]);
  });
});
