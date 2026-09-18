import { describe, it, expect } from 'vitest';
import { sortRows, toggleSort, paginateRows, pageCount, pageRange } from './table';

describe('sortRows', () => {
  const rows = [
    { name: 'Charlie', revenue: 300 },
    { name: 'Alice', revenue: 100 },
    { name: 'Bob', revenue: 200 },
  ];

  it('sorts strings ascending by key', () => {
    const result = sortRows(rows, { key: 'name', direction: 'asc' });
    expect(result.map((r) => r.name)).toEqual(['Alice', 'Bob', 'Charlie']);
  });

  it('sorts strings descending by key', () => {
    const result = sortRows(rows, { key: 'name', direction: 'desc' });
    expect(result[0].name).toBe('Charlie');
  });

  it('sorts numbers ascending', () => {
    const result = sortRows(rows, { key: 'revenue', direction: 'asc' });
    expect(result.map((r) => r.revenue)).toEqual([100, 200, 300]);
  });

  it('sorts numbers descending', () => {
    const result = sortRows(rows, { key: 'revenue', direction: 'desc' });
    expect(result[0].revenue).toBe(300);
  });

  it('does not mutate the original array', () => {
    sortRows(rows, { key: 'name', direction: 'asc' });
    expect(rows[0].name).toBe('Charlie');
  });
});

describe('toggleSort', () => {
  it('returns asc for a new key', () => {
    const result = toggleSort(null, 'revenue');
    expect(result).toEqual({ key: 'revenue', direction: 'asc' });
  });

  it('flips asc to desc for the same key', () => {
    const result = toggleSort({ key: 'revenue', direction: 'asc' }, 'revenue');
    expect(result.direction).toBe('desc');
  });

  it('flips desc to asc for the same key', () => {
    const result = toggleSort({ key: 'revenue', direction: 'desc' }, 'revenue');
    expect(result.direction).toBe('asc');
  });

  it('resets to asc when switching to a different key', () => {
    const result = toggleSort({ key: 'revenue', direction: 'desc' }, 'name');
    expect(result).toEqual({ key: 'name', direction: 'asc' });
  });
});

describe('paginateRows', () => {
  const rows = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10];

  it('returns the first page correctly', () => {
    expect(paginateRows(rows, 1, 3)).toEqual([1, 2, 3]);
  });

  it('returns the second page correctly', () => {
    expect(paginateRows(rows, 2, 3)).toEqual([4, 5, 6]);
  });

  it('returns a partial last page', () => {
    expect(paginateRows(rows, 4, 3)).toEqual([10]);
  });

  it('returns empty array for out-of-range page', () => {
    expect(paginateRows(rows, 10, 3)).toEqual([]);
  });

  it('returns empty array for page < 1', () => {
    expect(paginateRows(rows, 0, 3)).toEqual([]);
  });

  it('returns empty array for pageSize <= 0', () => {
    expect(paginateRows(rows, 1, 0)).toEqual([]);
  });
});

describe('pageCount', () => {
  it('calculates the correct number of pages', () => {
    expect(pageCount(100, 25)).toBe(4);
  });

  it('rounds up for partial pages', () => {
    expect(pageCount(101, 25)).toBe(5);
  });

  it('returns 0 for rowCount 0', () => {
    expect(pageCount(0, 25)).toBe(0);
  });

  it('returns 0 for pageSize 0', () => {
    expect(pageCount(100, 0)).toBe(0);
  });
});

describe('pageRange', () => {
  it('returns a range string', () => {
    expect(pageRange(100, 1, 25)).toBe('1-25 of 100');
  });

  it('caps end at rowCount on the last page', () => {
    expect(pageRange(102, 5, 25)).toBe('101-102 of 102');
  });

  it('returns "0 results" when rowCount is 0', () => {
    expect(pageRange(0, 1, 25)).toBe('0 results');
  });
});
