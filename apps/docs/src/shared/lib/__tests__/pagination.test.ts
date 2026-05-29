import { describe, it, expect } from 'vitest';
import { totalPages, paginate, hasNextPage, hasPrevPage, pageNumbers } from '../pagination';

describe('totalPages', () => {
  it('calculates total pages', () => {
    expect(totalPages(10, 3)).toBe(4);
  });

  it('returns 0 for invalid page size', () => {
    expect(totalPages(10, 0)).toBe(0);
  });

  it('returns 1 when items fit in one page', () => {
    expect(totalPages(5, 10)).toBe(1);
  });
});

describe('paginate', () => {
  const items = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10];

  it('returns first page', () => {
    expect(paginate(items, 1, 3)).toEqual([1, 2, 3]);
  });

  it('returns second page', () => {
    expect(paginate(items, 2, 3)).toEqual([4, 5, 6]);
  });

  it('returns partial last page', () => {
    expect(paginate(items, 4, 3)).toEqual([10]);
  });

  it('returns empty for invalid page', () => {
    expect(paginate(items, 0, 3)).toEqual([]);
  });
});

describe('hasNextPage', () => {
  it('returns true when not on last page', () => {
    expect(hasNextPage(1, 5)).toBe(true);
  });

  it('returns false on last page', () => {
    expect(hasNextPage(5, 5)).toBe(false);
  });
});

describe('hasPrevPage', () => {
  it('returns false on first page', () => {
    expect(hasPrevPage(1)).toBe(false);
  });

  it('returns true on later pages', () => {
    expect(hasPrevPage(2)).toBe(true);
  });
});

describe('pageNumbers', () => {
  it('always includes first and last', () => {
    const pages = pageNumbers(5, 10, 1);
    expect(pages).toContain(1);
    expect(pages).toContain(10);
  });

  it('includes pages around current', () => {
    const pages = pageNumbers(5, 10, 1);
    expect(pages).toContain(4);
    expect(pages).toContain(5);
    expect(pages).toContain(6);
  });

  it('returns sorted array', () => {
    const pages = pageNumbers(3, 5, 1);
    expect(pages).toEqual([...pages].sort((a, b) => a - b));
  });
});
