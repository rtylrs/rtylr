import { describe, it, expect } from 'vitest';

import {
  paginateItems,
  totalPages,
  hasNextPage,
  hasPrevPage,
  clampPage,
  pageRange,
  pageNumbers,
} from './paginationUtils';

describe('paginateItems', () => {
  const items = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10];

  it('returns first page items', () => {
    expect(paginateItems(items, 1, 3)).toEqual([1, 2, 3]);
  });

  it('returns second page items', () => {
    expect(paginateItems(items, 2, 3)).toEqual([4, 5, 6]);
  });

  it('returns partial last page', () => {
    expect(paginateItems(items, 4, 3)).toEqual([10]);
  });
});

describe('totalPages', () => {
  it('calculates total pages', () => {
    expect(totalPages(10, 3)).toBe(4);
    expect(totalPages(9, 3)).toBe(3);
    expect(totalPages(0, 3)).toBe(0);
  });

  it('returns 0 for zero page size', () => {
    expect(totalPages(10, 0)).toBe(0);
  });
});

describe('hasNextPage', () => {
  it('returns true when more pages exist', () => {
    expect(hasNextPage(1, 10, 3)).toBe(true);
  });

  it('returns false on last page', () => {
    expect(hasNextPage(4, 10, 3)).toBe(false);
  });
});

describe('hasPrevPage', () => {
  it('returns false on first page', () => {
    expect(hasPrevPage(1)).toBe(false);
  });

  it('returns true on page 2+', () => {
    expect(hasPrevPage(2)).toBe(true);
  });
});

describe('clampPage', () => {
  it('clamps below 1 to 1', () => {
    expect(clampPage(0, 10, 3)).toBe(1);
  });

  it('clamps above max to max', () => {
    expect(clampPage(99, 10, 3)).toBe(4);
  });

  it('returns 1 when no items', () => {
    expect(clampPage(5, 0, 3)).toBe(1);
  });
});

describe('pageRange', () => {
  it('returns correct range for first page', () => {
    expect(pageRange(1, 25, 10)).toEqual({ start: 1, end: 10, total: 25 });
  });

  it('returns correct range for last partial page', () => {
    expect(pageRange(3, 25, 10)).toEqual({ start: 21, end: 25, total: 25 });
  });

  it('returns zeros for empty list', () => {
    expect(pageRange(1, 0, 10)).toEqual({ start: 0, end: 0, total: 0 });
  });
});

describe('pageNumbers', () => {
  it('returns all pages when few total', () => {
    expect(pageNumbers(1, 3)).toEqual([1, 2, 3]);
  });

  it('inserts ellipsis (-1) for gaps', () => {
    const nums = pageNumbers(5, 10, 1);
    expect(nums).toContain(-1);
    expect(nums).toContain(1);
    expect(nums).toContain(10);
  });

  it('returns empty for 0 total pages', () => {
    expect(pageNumbers(1, 0)).toEqual([]);
  });
});
