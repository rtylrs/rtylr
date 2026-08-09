/**
 * Pagination utility functions used across ERP list views.
 */

export interface PaginationState {
  page: number;
  pageSize: number;
  totalItems: number;
}

/** Return the total number of pages for a given item count and page size. */
export function totalPages(totalItems: number, pageSize: number): number {
  if (pageSize <= 0) return 0;
  return Math.ceil(totalItems / pageSize);
}

/** Return true when there is a next page. */
export function hasNextPage(state: PaginationState): boolean {
  return state.page < totalPages(state.totalItems, state.pageSize);
}

/** Return true when there is a previous page. */
export function hasPrevPage(state: PaginationState): boolean {
  return state.page > 1;
}

/** Return the offset (zero-based index of the first item on the current page). */
export function pageOffset(page: number, pageSize: number): number {
  return (page - 1) * pageSize;
}

/** Return the 1-based range [first, last] of items shown on the current page. */
export function pageRange(page: number, pageSize: number, totalItems: number): [number, number] {
  const first = pageOffset(page, pageSize) + 1;
  const last = Math.min(page * pageSize, totalItems);
  return [first, last];
}

/** Return a human-readable range label such as "1-25 of 100". */
export function pageRangeLabel(page: number, pageSize: number, totalItems: number): string {
  const [first, last] = pageRange(page, pageSize, totalItems);
  return `${first}-${last} of ${totalItems}`;
}

/** Return an array of page numbers for pagination controls. */
export function pageNumbers(
  currentPage: number,
  total: number,
  maxVisible = 5
): (number | 'ellipsis')[] {
  if (total <= maxVisible) {
    return Array.from({ length: total }, (_, i) => i + 1);
  }
  const pages: (number | 'ellipsis')[] = [];
  const half = Math.floor(maxVisible / 2);
  let start = Math.max(1, currentPage - half);
  const end = Math.min(total, start + maxVisible - 1);
  if (end - start < maxVisible - 1) start = Math.max(1, end - maxVisible + 1);
  if (start > 1) {
    pages.push(1);
    if (start > 2) pages.push('ellipsis');
  }
  for (let i = start; i <= end; i++) pages.push(i);
  if (end < total) {
    if (end < total - 1) pages.push('ellipsis');
    pages.push(total);
  }
  return pages;
}
