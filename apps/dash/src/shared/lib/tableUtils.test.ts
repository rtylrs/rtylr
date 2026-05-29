import { describe, it, expect } from 'vitest';
import {
  toggleSort,
  sortRows,
  filterRows,
  pageRows,
  totalPages,
  visiblePages,
  rowRangeLabel,
} from './tableUtils';

describe('toggleSort', () => {
  it('sets a new key with ascending direction', () => {
    expect(toggleSort(null, 'name')).toEqual({ key: 'name', direction: 'asc' });
  });

  it('toggles direction when the same key is clicked', () => {
    expect(toggleSort({ key: 'name', direction: 'asc' }, 'name')).toEqual({
      key: 'name',
      direction: 'desc',
    });
  });

  it('resets to asc when a different key is clicked', () => {
    expect(toggleSort({ key: 'name', direction: 'desc' }, 'age')).toEqual({
      key: 'age',
      direction: 'asc',
    });
  });
});

describe('sortRows', () => {
  const rows = [
    { id: 1, name: 'Charlie', age: 30 },
    { id: 2, name: 'Alice', age: 25 },
    { id: 3, name: 'Bob', age: 35 },
  ];

  it('sorts by string field ascending', () => {
    const sorted = sortRows(rows, { key: 'name', direction: 'asc' });
    expect(sorted.map((r) => r.name)).toEqual(['Alice', 'Bob', 'Charlie']);
  });

  it('sorts by number field descending', () => {
    const sorted = sortRows(rows, { key: 'age', direction: 'desc' });
    expect(sorted.map((r) => r.age)).toEqual([35, 30, 25]);
  });

  it('returns the original order when sort is null', () => {
    expect(sortRows(rows, null)).toEqual(rows);
  });

  it('does not mutate the original array', () => {
    const original = [...rows];
    sortRows(rows, { key: 'name', direction: 'asc' });
    expect(rows).toEqual(original);
  });
});

describe('filterRows', () => {
  const rows = [
    { id: 1, name: 'Alice Smith', city: 'London' },
    { id: 2, name: 'Bob Jones', city: 'Paris' },
    { id: 3, name: 'Carol White', city: 'London' },
  ];

  it('returns all rows for an empty query', () => {
    expect(filterRows(rows, '')).toHaveLength(3);
  });

  it('filters rows by a string field', () => {
    expect(filterRows(rows, 'alice')).toHaveLength(1);
  });

  it('is case-insensitive', () => {
    expect(filterRows(rows, 'LONDON')).toHaveLength(2);
  });

  it('filters only specified fields', () => {
    const result = filterRows(rows, 'jones', ['name']);
    expect(result).toHaveLength(1);
    expect(result[0]!.name).toBe('Bob Jones');
  });
});

describe('pageRows', () => {
  const rows = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10];

  it('returns the first page', () => {
    expect(pageRows(rows, 1, 3)).toEqual([1, 2, 3]);
  });

  it('returns the second page', () => {
    expect(pageRows(rows, 2, 3)).toEqual([4, 5, 6]);
  });

  it('returns a partial last page', () => {
    expect(pageRows(rows, 4, 3)).toEqual([10]);
  });
});

describe('totalPages', () => {
  it('returns correct total pages', () => {
    expect(totalPages(100, 25)).toBe(4);
  });

  it('rounds up for partial pages', () => {
    expect(totalPages(101, 25)).toBe(5);
  });

  it('returns 0 for empty set', () => {
    expect(totalPages(0, 25)).toBe(0);
  });
});

describe('visiblePages', () => {
  it('returns all pages when total is within window', () => {
    expect(visiblePages(1, 3, 5)).toEqual([1, 2, 3]);
  });

  it('centers around the current page', () => {
    expect(visiblePages(5, 10, 5)).toEqual([3, 4, 5, 6, 7]);
  });

  it('clamps to the start', () => {
    expect(visiblePages(1, 10, 5)).toEqual([1, 2, 3, 4, 5]);
  });

  it('clamps to the end', () => {
    expect(visiblePages(10, 10, 5)).toEqual([6, 7, 8, 9, 10]);
  });

  it('returns empty array when total is 0', () => {
    expect(visiblePages(1, 0)).toEqual([]);
  });
});

describe('rowRangeLabel', () => {
  it('returns "1-25 of 100" for page 1 size 25 total 100', () => {
    expect(rowRangeLabel(1, 25, 100)).toBe('1-25 of 100');
  });

  it('returns "26-50 of 100" for page 2', () => {
    expect(rowRangeLabel(2, 25, 100)).toBe('26-50 of 100');
  });

  it('handles partial last page', () => {
    expect(rowRangeLabel(4, 25, 90)).toBe('76-90 of 90');
  });

  it('returns "0 results" for empty dataset', () => {
    expect(rowRangeLabel(1, 25, 0)).toBe('0 results');
  });
});
