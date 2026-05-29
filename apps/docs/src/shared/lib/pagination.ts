/** Pagination helpers for the docs app. */

/**
 * Returns the total number of pages for the given item count and page size.
 */
export function totalPages(itemCount: number, pageSize: number): number {
  if (pageSize <= 0) return 0;
  return Math.ceil(itemCount / pageSize);
}

/**
 * Returns the slice of items for the given 1-based page number.
 */
export function paginate<T>(items: T[], page: number, pageSize: number): T[] {
  if (pageSize <= 0 || page < 1) return [];
  const start = (page - 1) * pageSize;
  return items.slice(start, start + pageSize);
}

/**
 * Returns true if there is a next page.
 */
export function hasNextPage(currentPage: number, total: number): boolean {
  return currentPage < total;
}

/**
 * Returns true if there is a previous page.
 */
export function hasPrevPage(currentPage: number): boolean {
  return currentPage > 1;
}

/**
 * Returns an array of page numbers to render in a pagination control.
 * Always includes first, last, and up to `window` pages around the current page.
 */
export function pageNumbers(current: number, total: number, window = 2): number[] {
  const pages = new Set<number>();
  pages.add(1);
  pages.add(total);
  for (let i = Math.max(1, current - window); i <= Math.min(total, current + window); i++) {
    pages.add(i);
  }
  return Array.from(pages).sort((a, b) => a - b);
}
