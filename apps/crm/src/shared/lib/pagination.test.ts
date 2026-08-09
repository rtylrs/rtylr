import { describe, it, expect } from 'vitest';
import { getPaginationInfo, clampPage, getPageNumbers } from './pagination';

// ── getPaginationInfo ─────────────────────────────────────────────────────────

describe('getPaginationInfo', () => {
  it('returns correct metadata for the first page', () => {
    const info = getPaginationInfo({ page: 1, pageSize: 10, total: 35 });
    expect(info.totalPages).toBe(4);
    expect(info.offset).toBe(0);
    expect(info.hasPrev).toBe(false);
    expect(info.hasNext).toBe(true);
    expect(info.startIndex).toBe(1);
    expect(info.endIndex).toBe(10);
  });

  it('returns correct metadata for the last page', () => {
    const info = getPaginationInfo({ page: 4, pageSize: 10, total: 35 });
    expect(info.offset).toBe(30);
    expect(info.hasPrev).toBe(true);
    expect(info.hasNext).toBe(false);
    expect(info.startIndex).toBe(31);
    expect(info.endIndex).toBe(35);
  });

  it('returns correct metadata for a middle page', () => {
    const info = getPaginationInfo({ page: 2, pageSize: 10, total: 35 });
    expect(info.offset).toBe(10);
    expect(info.hasPrev).toBe(true);
    expect(info.hasNext).toBe(true);
    expect(info.startIndex).toBe(11);
    expect(info.endIndex).toBe(20);
  });

  it('handles zero total items', () => {
    const info = getPaginationInfo({ page: 1, pageSize: 10, total: 0 });
    expect(info.totalPages).toBe(0);
    expect(info.startIndex).toBe(0);
    expect(info.endIndex).toBe(0);
    expect(info.hasPrev).toBe(false);
    expect(info.hasNext).toBe(false);
  });

  it('clamps page number that exceeds totalPages', () => {
    const info = getPaginationInfo({ page: 99, pageSize: 10, total: 5 });
    expect(info.page).toBe(1);
  });

  it('handles a total that exactly fills one page', () => {
    const info = getPaginationInfo({ page: 1, pageSize: 5, total: 5 });
    expect(info.totalPages).toBe(1);
    expect(info.hasNext).toBe(false);
    expect(info.hasPrev).toBe(false);
  });
});

// ── clampPage ─────────────────────────────────────────────────────────────────

describe('clampPage', () => {
  it('returns 1 for page 0', () => {
    expect(clampPage(0, 5)).toBe(1);
  });

  it('returns 1 for negative page', () => {
    expect(clampPage(-3, 5)).toBe(1);
  });

  it('returns totalPages for page exceeding totalPages', () => {
    expect(clampPage(10, 5)).toBe(5);
  });

  it('returns the page when it is within bounds', () => {
    expect(clampPage(3, 5)).toBe(3);
  });

  it('returns 1 when totalPages is 0', () => {
    expect(clampPage(1, 0)).toBe(1);
  });
});

// ── getPageNumbers ────────────────────────────────────────────────────────────

describe('getPageNumbers', () => {
  it('returns an empty array when totalPages is 0', () => {
    expect(getPageNumbers(1, 0)).toEqual([]);
  });

  it('returns [1] when totalPages is 1', () => {
    expect(getPageNumbers(1, 1)).toEqual([1]);
  });

  it('returns all pages when totalPages is small', () => {
    expect(getPageNumbers(3, 5)).toEqual([1, 2, 3, 4, 5]);
  });

  it('inserts a gap (-1) between non-adjacent page blocks', () => {
    const pages = getPageNumbers(1, 10, 1);
    expect(pages).toContain(-1);
    expect(pages[0]).toBe(1);
    expect(pages[pages.length - 1]).toBe(10);
  });

  it('always includes page 1 and the last page', () => {
    const pages = getPageNumbers(5, 20, 2);
    expect(pages[0]).toBe(1);
    expect(pages[pages.length - 1]).toBe(20);
  });
});
