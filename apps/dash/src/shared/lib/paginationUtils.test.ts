import { describe, it, expect } from 'vitest';
import {
  totalPages,
  hasPrevPage,
  hasNextPage,
  firstItemIndex,
  lastItemIndex,
  paginationLabel,
  pageNumbers,
} from './paginationUtils';

describe('totalPages', () => {
  it('computes correct page count', () => {
    expect(totalPages(100, 10)).toBe(10);
  });

  it('rounds up for non-divisible totals', () => {
    expect(totalPages(101, 10)).toBe(11);
  });

  it('returns 0 for total of 0', () => {
    expect(totalPages(0, 10)).toBe(0);
  });

  it('returns 0 for pageSize of 0', () => {
    expect(totalPages(100, 0)).toBe(0);
  });
});

describe('hasPrevPage', () => {
  it('returns false for page 1', () => {
    expect(hasPrevPage(1)).toBe(false);
  });

  it('returns true for page 2', () => {
    expect(hasPrevPage(2)).toBe(true);
  });
});

describe('hasNextPage', () => {
  it('returns true when on first of multiple pages', () => {
    expect(hasNextPage(1, 100, 10)).toBe(true);
  });

  it('returns false when on the last page', () => {
    expect(hasNextPage(10, 100, 10)).toBe(false);
  });

  it('returns false when on the only page', () => {
    expect(hasNextPage(1, 5, 10)).toBe(false);
  });
});

describe('firstItemIndex', () => {
  it('returns 1 for page 1', () => {
    expect(firstItemIndex(1, 10)).toBe(1);
  });

  it('returns 11 for page 2 with pageSize 10', () => {
    expect(firstItemIndex(2, 10)).toBe(11);
  });
});

describe('lastItemIndex', () => {
  it('returns pageSize for a full first page', () => {
    expect(lastItemIndex(1, 10, 100)).toBe(10);
  });

  it('returns total for the last partial page', () => {
    expect(lastItemIndex(11, 10, 105)).toBe(105);
  });
});

describe('paginationLabel', () => {
  it('produces correct label for first page', () => {
    expect(paginationLabel({ page: 1, pageSize: 10, total: 100 })).toBe(
      'Showing 1 - 10 of 100'
    );
  });

  it('produces correct label for partial last page', () => {
    expect(paginationLabel({ page: 11, pageSize: 10, total: 105 })).toBe(
      'Showing 101 - 105 of 105'
    );
  });

  it('returns "No results" for total of 0', () => {
    expect(paginationLabel({ page: 1, pageSize: 10, total: 0 })).toBe('No results');
  });
});

describe('pageNumbers', () => {
  it('includes first and last page', () => {
    const pages = pageNumbers(5, 10);
    expect(pages).toContain(1);
    expect(pages).toContain(10);
  });

  it('includes current page and neighbors', () => {
    const pages = pageNumbers(5, 10, 1);
    expect(pages).toContain(4);
    expect(pages).toContain(5);
    expect(pages).toContain(6);
  });

  it('includes 0 as ellipsis placeholder for gaps', () => {
    const pages = pageNumbers(5, 20, 1);
    expect(pages).toContain(0);
  });

  it('returns empty array for 0 pages', () => {
    expect(pageNumbers(1, 0)).toEqual([]);
  });

  it('returns single page for 1 total page', () => {
    expect(pageNumbers(1, 1)).toEqual([1]);
  });
});
