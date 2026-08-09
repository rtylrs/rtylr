import { describe, it, expect } from 'vitest';
import {
  compareValues,
  compareBy,
  compareByMultiple,
  sortWith,
  toggleSortDir,
  nextSortState,
} from './sortUtils';

describe('compareValues', () => {
  it('returns 0 for equal values', () => {
    expect(compareValues(5, 5)).toBe(0);
  });

  it('returns negative for a < b', () => {
    expect(compareValues(1, 2)).toBeLessThan(0);
  });

  it('returns positive for a > b', () => {
    expect(compareValues('b', 'a')).toBeGreaterThan(0);
  });

  it('puts nulls last', () => {
    expect(compareValues(null, 5)).toBeGreaterThan(0);
    expect(compareValues(5, null)).toBeLessThan(0);
  });

  it('compares strings with localeCompare', () => {
    expect(compareValues('apple', 'banana')).toBeLessThan(0);
  });
});

describe('compareBy', () => {
  const items = [
    { name: 'Zebra', n: 3 },
    { name: 'Apple', n: 1 },
  ];

  it('sorts by key ascending', () => {
    const sorted = [...items].sort(compareBy('name'));
    expect(sorted[0].name).toBe('Apple');
  });

  it('sorts by key descending', () => {
    const sorted = [...items].sort(compareBy('n', 'desc'));
    expect(sorted[0].n).toBe(3);
  });
});

describe('compareByMultiple', () => {
  const items = [
    { cat: 'A', name: 'Zebra' },
    { cat: 'B', name: 'Apple' },
    { cat: 'A', name: 'Apple' },
  ];

  it('sorts by multiple keys', () => {
    const sorted = [...items].sort(compareByMultiple([{ key: 'cat' }, { key: 'name' }]));
    expect(sorted[0]).toEqual({ cat: 'A', name: 'Apple' });
    expect(sorted[1]).toEqual({ cat: 'A', name: 'Zebra' });
  });
});

describe('sortWith', () => {
  it('returns sorted copy without mutating original', () => {
    const items = [3, 1, 2];
    const sorted = sortWith(items, (a, b) => a - b);
    expect(sorted).toEqual([1, 2, 3]);
    expect(items).toEqual([3, 1, 2]);
  });
});

describe('toggleSortDir', () => {
  it('toggles asc to desc', () => {
    expect(toggleSortDir('asc')).toBe('desc');
  });

  it('toggles desc to asc', () => {
    expect(toggleSortDir('desc')).toBe('asc');
  });
});

describe('nextSortState', () => {
  it('sets asc when clicking a new column', () => {
    const result = nextSortState('name', { column: null, dir: 'asc' });
    expect(result).toEqual({ column: 'name', dir: 'asc' });
  });

  it('toggles direction when clicking same column', () => {
    const result = nextSortState('name', { column: 'name', dir: 'asc' });
    expect(result).toEqual({ column: 'name', dir: 'desc' });
  });
});
