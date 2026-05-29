import { describe, it, expect } from 'vitest';
import { chunk, take, takeLast, sortBy, partition, flatten, unique, zip } from './arrayUtils';

describe('chunk', () => {
  it('splits into equal chunks', () =>
    expect(chunk([1, 2, 3, 4], 2)).toEqual([
      [1, 2],
      [3, 4],
    ]));
  it('handles remainder', () => expect(chunk([1, 2, 3], 2)).toEqual([[1, 2], [3]]));
  it('returns [] for size 0', () => expect(chunk([1, 2], 0)).toEqual([]));
});

describe('take', () => {
  it('returns first n items', () => expect(take([1, 2, 3, 4], 2)).toEqual([1, 2]));
  it('returns all when n > length', () => expect(take([1, 2], 10)).toEqual([1, 2]));
});

describe('takeLast', () => {
  it('returns last n items', () => expect(takeLast([1, 2, 3, 4], 2)).toEqual([3, 4]));
  it('returns all when n > length', () => expect(takeLast([1, 2], 10)).toEqual([1, 2]));
});

describe('sortBy', () => {
  it('sorts by key', () => {
    const arr = [{ n: 3 }, { n: 1 }, { n: 2 }];
    expect(sortBy(arr, (x) => x.n)).toEqual([{ n: 1 }, { n: 2 }, { n: 3 }]);
  });
  it('does not mutate', () => {
    const arr = [3, 1, 2];
    sortBy(arr, (x) => x);
    expect(arr).toEqual([3, 1, 2]);
  });
});

describe('partition', () => {
  it('splits on predicate', () => {
    const [evens, odds] = partition([1, 2, 3, 4], (n) => n % 2 === 0);
    expect(evens).toEqual([2, 4]);
    expect(odds).toEqual([1, 3]);
  });
});

describe('flatten', () => {
  it('flattens 2D array', () =>
    expect(
      flatten([
        [1, 2],
        [3, 4],
      ])
    ).toEqual([1, 2, 3, 4]));
});

describe('unique', () => {
  it('removes duplicates', () => expect(unique([1, 2, 1, 3])).toEqual([1, 2, 3]));
  it('handles strings', () => expect(unique(['a', 'b', 'a'])).toEqual(['a', 'b']));
});

describe('zip', () => {
  it('pairs elements', () =>
    expect(zip([1, 2], ['a', 'b'])).toEqual([
      [1, 'a'],
      [2, 'b'],
    ]));
  it('truncates to shorter', () => expect(zip([1, 2, 3], ['a'])).toEqual([[1, 'a']]));
});
