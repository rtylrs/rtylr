// Pagination helpers for the insights feature.

export interface PaginationState {
  page: number;
  pageSize: number;
  total: number;
}

/**
 * totalPages returns the total number of pages given total items and page size.
 */
export function totalPages(total: number, pageSize: number): number {
  if (pageSize <= 0) return 0;
  return Math.ceil(total / pageSize);
}

/**
 * pageSlice returns the start and end indices for the current page.
 */
export function pageSlice(page: number, pageSize: number): { start: number; end: number } {
  const start = (page - 1) * pageSize;
  return { start, end: start + pageSize };
}

/**
 * paginateItems returns the subset of items for the current page.
 */
export function paginateItems<T>(items: T[], page: number, pageSize: number): T[] {
  const { start, end } = pageSlice(page, pageSize);
  return items.slice(start, end);
}

/**
 * hasNextPage returns true when there is a page after the current one.
 */
export function hasNextPage(state: PaginationState): boolean {
  return state.page < totalPages(state.total, state.pageSize);
}

/**
 * hasPrevPage returns true when there is a page before the current one.
 */
export function hasPrevPage(state: PaginationState): boolean {
  return state.page > 1;
}

/**
 * pageRange returns an array of page numbers for pagination UI.
 * Always includes first and last pages, with gaps represented as -1.
 */
export function pageRange(current: number, total: number, delta = 2): number[] {
  if (total <= 1) return total === 1 ? [1] : [];
  const pages = new Set<number>();
  pages.add(1);
  pages.add(total);
  for (let i = Math.max(2, current - delta); i <= Math.min(total - 1, current + delta); i++) {
    pages.add(i);
  }
  const sorted = [...pages].sort((a, b) => a - b);
  const result: number[] = [];
  for (let i = 0; i < sorted.length; i++) {
    result.push(sorted[i]);
    if (i < sorted.length - 1 && sorted[i + 1] - sorted[i] > 1) {
      result.push(-1);
    }
  }
  return result;
}
