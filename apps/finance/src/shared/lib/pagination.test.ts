import { describe, it, expect } from 'vitest';
import {
  getPaginationMeta,
  hasPreviousPage,
  hasNextPage,
  clampPage,
  getPageStartIndex,
  getPageEndIndex,
} from './pagination';

describe('getPaginationMeta', () => {
  it('computes total pages correctly', () => {
    const meta = getPaginationMeta(1, 10, 25);
    expect(meta.totalPages).toBe(3);
  });

  it('reports hasNext when not on last page', () => {
    const meta = getPaginationMeta(1, 10, 25);
    expect(meta.hasNext).toBe(true);
  });

  it('reports no hasNext on last page', () => {
    const meta = getPaginationMeta(3, 10, 25);
    expect(meta.hasNext).toBe(false);
  });

  it('reports hasPrevious on page 2+', () => {
    const meta = getPaginationMeta(2, 10, 25);
    expect(meta.hasPrevious).toBe(true);
  });

  it('reports no hasPrevious on page 1', () => {
    const meta = getPaginationMeta(1, 10, 25);
    expect(meta.hasPrevious).toBe(false);
  });

  it('returns totalPages 0 when pageSize is 0', () => {
    const meta = getPaginationMeta(1, 0, 100);
    expect(meta.totalPages).toBe(0);
  });

  it('returns all meta fields', () => {
    const meta = getPaginationMeta(2, 20, 50);
    expect(meta.page).toBe(2);
    expect(meta.pageSize).toBe(20);
    expect(meta.totalCount).toBe(50);
  });
});

describe('hasPreviousPage', () => {
  it('returns false on page 1', () => {
    expect(hasPreviousPage(1)).toBe(false);
  });

  it('returns true on page 2', () => {
    expect(hasPreviousPage(2)).toBe(true);
  });
});

describe('hasNextPage', () => {
  it('returns true when page < totalPages', () => {
    expect(hasNextPage(1, 3)).toBe(true);
  });

  it('returns false when page equals totalPages', () => {
    expect(hasNextPage(3, 3)).toBe(false);
  });
});

describe('clampPage', () => {
  it('clamps below 1 to 1', () => {
    expect(clampPage(0, 5)).toBe(1);
    expect(clampPage(-3, 5)).toBe(1);
  });

  it('clamps above totalPages to totalPages', () => {
    expect(clampPage(10, 5)).toBe(5);
  });

  it('returns page unchanged when within range', () => {
    expect(clampPage(3, 5)).toBe(3);
  });

  it('returns 1 when totalPages is 0', () => {
    expect(clampPage(5, 0)).toBe(1);
  });
});

describe('getPageStartIndex', () => {
  it('returns 0 for first page', () => {
    expect(getPageStartIndex(1, 10)).toBe(0);
  });

  it('returns correct offset for page 2', () => {
    expect(getPageStartIndex(2, 10)).toBe(10);
  });

  it('returns correct offset for page 3', () => {
    expect(getPageStartIndex(3, 25)).toBe(50);
  });
});

describe('getPageEndIndex', () => {
  it('returns pageSize for first full page', () => {
    expect(getPageEndIndex(1, 10, 100)).toBe(10);
  });

  it('returns totalCount for partial last page', () => {
    expect(getPageEndIndex(3, 10, 25)).toBe(25);
  });

  it('returns full page size when exactly aligned', () => {
    expect(getPageEndIndex(2, 10, 20)).toBe(20);
  });
});
