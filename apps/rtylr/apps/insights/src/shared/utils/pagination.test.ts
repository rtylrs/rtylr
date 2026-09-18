import { describe, it, expect } from 'vitest';
import {
  totalPages,
  pageSlice,
  paginateItems,
  hasNextPage,
  hasPrevPage,
  pageRange,
} from './pagination';

describe('totalPages', () => {
  it('returns correct page count', () => {
    expect(totalPages(100, 10)).toBe(10);
    expect(totalPages(11, 10)).toBe(2);
  });

  it('returns 0 for invalid pageSize', () => {
    expect(totalPages(100, 0)).toBe(0);
  });
});

describe('pageSlice', () => {
  it('returns correct start and end for page 1', () => {
    expect(pageSlice(1, 10)).toEqual({ start: 0, end: 10 });
  });

  it('returns correct indices for page 2', () => {
    expect(pageSlice(2, 10)).toEqual({ start: 10, end: 20 });
  });
});

describe('paginateItems', () => {
  it('returns the correct page subset', () => {
    const items = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10];
    expect(paginateItems(items, 1, 3)).toEqual([1, 2, 3]);
    expect(paginateItems(items, 2, 3)).toEqual([4, 5, 6]);
  });

  it('returns partial page at the end', () => {
    const items = [1, 2, 3, 4, 5];
    expect(paginateItems(items, 2, 3)).toEqual([4, 5]);
  });
});

describe('hasNextPage', () => {
  it('returns true when not on last page', () => {
    expect(hasNextPage({ page: 1, pageSize: 10, total: 25 })).toBe(true);
  });

  it('returns false when on last page', () => {
    expect(hasNextPage({ page: 3, pageSize: 10, total: 25 })).toBe(false);
  });
});

describe('hasPrevPage', () => {
  it('returns false on first page', () => {
    expect(hasPrevPage({ page: 1, pageSize: 10, total: 25 })).toBe(false);
  });

  it('returns true when not on first page', () => {
    expect(hasPrevPage({ page: 2, pageSize: 10, total: 25 })).toBe(true);
  });
});

describe('pageRange', () => {
  it('returns all pages for small total', () => {
    expect(pageRange(1, 3)).toEqual([1, 2, 3]);
  });

  it('inserts -1 for gaps', () => {
    const range = pageRange(5, 10, 1);
    expect(range).toContain(-1);
    expect(range).toContain(1);
    expect(range).toContain(10);
  });

  it('returns empty for total 0', () => {
    expect(pageRange(1, 0)).toEqual([]);
  });
});
