import { describe, it, expect } from 'vitest';
import {
  totalPages,
  pageOffset,
  paginate,
  paginationResult,
  visiblePages,
  clampPage,
} from './paginationUtils';

describe('totalPages', () => {
  it('calculates total pages correctly', () => {
    expect(totalPages(10, 3)).toBe(4);
    expect(totalPages(9, 3)).toBe(3);
  });

  it('returns 0 for zero page size', () => {
    expect(totalPages(10, 0)).toBe(0);
  });

  it('returns 0 for zero total', () => {
    expect(totalPages(0, 10)).toBe(0);
  });
});

describe('pageOffset', () => {
  it('returns 0 for page 1', () => {
    expect(pageOffset(1, 10)).toBe(0);
  });

  it('calculates offset for subsequent pages', () => {
    expect(pageOffset(2, 10)).toBe(10);
    expect(pageOffset(3, 5)).toBe(10);
  });

  it('handles page 0 as page 1', () => {
    expect(pageOffset(0, 10)).toBe(0);
  });
});

describe('paginate', () => {
  const items = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10];

  it('returns the first page', () => {
    expect(paginate(items, 1, 3)).toEqual([1, 2, 3]);
  });

  it('returns the second page', () => {
    expect(paginate(items, 2, 3)).toEqual([4, 5, 6]);
  });

  it('returns partial last page', () => {
    expect(paginate(items, 4, 3)).toEqual([10]);
  });

  it('returns empty array for out-of-range page', () => {
    expect(paginate(items, 10, 3)).toEqual([]);
  });
});

describe('paginationResult', () => {
  const all = [1, 2, 3, 4, 5];

  it('returns correct metadata', () => {
    const page = paginate(all, 1, 2);
    const result = paginationResult(page, all, 1, 2);
    expect(result.total).toBe(5);
    expect(result.totalPages).toBe(3);
    expect(result.hasPrev).toBe(false);
    expect(result.hasNext).toBe(true);
  });

  it('sets hasPrev true on middle page', () => {
    const page = paginate(all, 2, 2);
    const result = paginationResult(page, all, 2, 2);
    expect(result.hasPrev).toBe(true);
    expect(result.hasNext).toBe(true);
  });
});

describe('visiblePages', () => {
  it('returns pages around current with delta 2', () => {
    expect(visiblePages(5, 10, 2)).toEqual([3, 4, 5, 6, 7]);
  });

  it('clamps to first page', () => {
    expect(visiblePages(1, 10, 2)).toEqual([1, 2, 3]);
  });

  it('clamps to last page', () => {
    expect(visiblePages(10, 10, 2)).toEqual([8, 9, 10]);
  });
});

describe('clampPage', () => {
  it('clamps to 1 when below range', () => {
    expect(clampPage(-1, 5)).toBe(1);
    expect(clampPage(0, 5)).toBe(1);
  });

  it('clamps to total when above range', () => {
    expect(clampPage(100, 5)).toBe(5);
  });

  it('returns valid page unchanged', () => {
    expect(clampPage(3, 5)).toBe(3);
  });

  it('returns 1 when total is 0', () => {
    expect(clampPage(5, 0)).toBe(1);
  });
});
