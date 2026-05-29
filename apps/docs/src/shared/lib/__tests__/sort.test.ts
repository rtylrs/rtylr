import { describe, it, expect } from 'vitest';
import { sortByTitle, sortByDate, sortByOrder, reverse } from '../sort';

const docs = [
  { title: 'Zebra', date: '2024-01-01', order: 3 },
  { title: 'Apple', date: '2024-06-01', order: 1 },
  { title: 'Mango', date: '2023-12-01', order: 2 },
];

describe('sortByTitle', () => {
  it('sorts alphabetically', () => {
    const sorted = sortByTitle(docs);
    expect(sorted[0].title).toBe('Apple');
    expect(sorted[2].title).toBe('Zebra');
  });

  it('does not mutate original', () => {
    const original = [...docs];
    sortByTitle(docs);
    expect(docs[0].title).toBe(original[0].title);
  });
});

describe('sortByDate', () => {
  it('sorts newest first', () => {
    const sorted = sortByDate(docs);
    expect(sorted[0].title).toBe('Apple');
  });

  it('puts oldest last', () => {
    const sorted = sortByDate(docs);
    expect(sorted[sorted.length - 1].title).toBe('Mango');
  });
});

describe('sortByOrder', () => {
  it('sorts by order field', () => {
    const sorted = sortByOrder(docs);
    expect(sorted[0].order).toBe(1);
    expect(sorted[2].order).toBe(3);
  });

  it('puts items without order last', () => {
    const mixed = [{ title: 'A' }, { title: 'B', order: 1 }];
    const sorted = sortByOrder(mixed);
    expect(sorted[0].title).toBe('B');
  });
});

describe('reverse', () => {
  it('reverses array', () => {
    expect(reverse([1, 2, 3])).toEqual([3, 2, 1]);
  });

  it('does not mutate original', () => {
    const arr = [1, 2, 3];
    reverse(arr);
    expect(arr).toEqual([1, 2, 3]);
  });
});
