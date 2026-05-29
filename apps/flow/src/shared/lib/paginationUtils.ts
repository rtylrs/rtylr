/** Returns the items for a given page (1-indexed). */
export function paginateItems<T>(items: T[], page: number, pageSize: number): T[] {
  const start = (page - 1) * pageSize;
  return items.slice(start, start + pageSize);
}

/** Returns the total number of pages for a given item count and page size. */
export function totalPages(count: number, pageSize: number): number {
  if (pageSize <= 0) return 0;
  return Math.ceil(count / pageSize);
}

/** Returns true if there is a next page. */
export function hasNextPage(page: number, count: number, pageSize: number): boolean {
  return page < totalPages(count, pageSize);
}

/** Returns true if there is a previous page. */
export function hasPrevPage(page: number): boolean {
  return page > 1;
}

/** Clamps a page number to valid range [1, totalPages]. */
export function clampPage(page: number, count: number, pageSize: number): number {
  const max = totalPages(count, pageSize);
  if (max === 0) return 1;
  return Math.min(Math.max(1, page), max);
}

/** Returns the start and end item indices (1-indexed) for display (e.g. "Showing 1-10 of 50"). */
export function pageRange(
  page: number,
  count: number,
  pageSize: number
): { start: number; end: number; total: number } {
  if (count === 0) return { start: 0, end: 0, total: 0 };
  const start = (page - 1) * pageSize + 1;
  const end = Math.min(page * pageSize, count);
  return { start, end, total: count };
}

/** Returns a list of page numbers to display in a paginator (with -1 for ellipsis). */
export function pageNumbers(currentPage: number, total: number, windowSize = 2): (number | -1)[] {
  const pages: (number | -1)[] = [];
  const max = total;
  if (max <= 1) return total === 1 ? [1] : [];

  // Always include first, last, and window around current
  const set = new Set<number>();
  set.add(1);
  set.add(max);
  for (let i = currentPage - windowSize; i <= currentPage + windowSize; i++) {
    if (i > 0 && i <= max) set.add(i);
  }

  const sorted = Array.from(set).sort((a, b) => a - b);
  for (let i = 0; i < sorted.length; i++) {
    pages.push(sorted[i]);
    if (i < sorted.length - 1 && sorted[i + 1] - sorted[i] > 1) {
      pages.push(-1);
    }
  }
  return pages;
}
