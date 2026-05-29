/**
 * Pagination helpers for the Menu app.
 */

export interface PaginationState {
  page: number;
  pageSize: number;
  total: number;
}

export interface PaginationResult<T> {
  items: T[];
  page: number;
  pageSize: number;
  total: number;
  totalPages: number;
  hasPrev: boolean;
  hasNext: boolean;
}

/** Return the total number of pages. */
export function totalPages(total: number, pageSize: number): number {
  if (pageSize <= 0) return 0;
  return Math.ceil(total / pageSize);
}

/** Return the offset (start index) for a given page. */
export function pageOffset(page: number, pageSize: number): number {
  return Math.max(0, page - 1) * pageSize;
}

/** Slice an array to a specific page. */
export function paginate<T>(items: T[], page: number, pageSize: number): T[] {
  const offset = pageOffset(page, pageSize);
  return items.slice(offset, offset + pageSize);
}

/** Return full pagination metadata for a result set. */
export function paginationResult<T>(
  items: T[],
  allItems: T[],
  page: number,
  pageSize: number
): PaginationResult<T> {
  const total = allItems.length;
  const pages = totalPages(total, pageSize);
  return {
    items,
    page,
    pageSize,
    total,
    totalPages: pages,
    hasPrev: page > 1,
    hasNext: page < pages,
  };
}

/** Return a range of visible page numbers around the current page. */
export function visiblePages(current: number, total: number, delta = 2): number[] {
  const start = Math.max(1, current - delta);
  const end = Math.min(total, current + delta);
  const pages: number[] = [];
  for (let i = start; i <= end; i++) pages.push(i);
  return pages;
}

/** Clamp a page number to valid bounds. */
export function clampPage(page: number, total: number): number {
  if (total <= 0) return 1;
  return Math.min(Math.max(1, page), total);
}
