/**
 * Pagination utility helpers for the Dash app.
 * All functions are pure with no side-effects.
 */

export interface PaginationState {
  page: number;
  pageSize: number;
  total: number;
}

/** Compute total number of pages. */
export function totalPages(total: number, pageSize: number): number {
  if (pageSize <= 0 || total <= 0) return 0;
  return Math.ceil(total / pageSize);
}

/** Return true when there is a previous page. */
export function hasPrevPage(page: number): boolean {
  return page > 1;
}

/** Return true when there is a next page. */
export function hasNextPage(page: number, total: number, pageSize: number): boolean {
  return page < totalPages(total, pageSize);
}

/** Compute the 1-based index of the first item on a page. */
export function firstItemIndex(page: number, pageSize: number): number {
  return (page - 1) * pageSize + 1;
}

/** Compute the 1-based index of the last item on a page. */
export function lastItemIndex(page: number, pageSize: number, total: number): number {
  return Math.min(page * pageSize, total);
}

/** Return a human-readable "Showing X - Y of Z" label. */
export function paginationLabel(state: PaginationState): string {
  if (state.total === 0) return 'No results';
  const first = firstItemIndex(state.page, state.pageSize);
  const last = lastItemIndex(state.page, state.pageSize, state.total);
  return `Showing ${first} - ${last} of ${state.total}`;
}

/**
 * Build a list of page numbers to display in a paginator.
 * Always includes first and last page; adds ellipsis placeholders (0) for gaps.
 */
export function pageNumbers(
  currentPage: number,
  pages: number,
  windowSize = 2
): (number | 0)[] {
  if (pages <= 0) return [];
  const set = new Set<number>();
  set.add(1);
  set.add(pages);
  for (
    let i = Math.max(1, currentPage - windowSize);
    i <= Math.min(pages, currentPage + windowSize);
    i++
  ) {
    set.add(i);
  }
  const sorted = [...set].sort((a, b) => a - b);
  const result: (number | 0)[] = [];
  for (let i = 0; i < sorted.length; i++) {
    if (i > 0 && sorted[i]! - sorted[i - 1]! > 1) result.push(0);
    result.push(sorted[i]!);
  }
  return result;
}
