// Table utilities for the insights app.
// Shared helpers for sorting, pagination, and column value extraction
// used by the data-table components in the features directory.

export type SortDirection = 'asc' | 'desc';

export interface SortState {
  key: string;
  direction: SortDirection;
}

/**
 * sortRows sorts an array of objects by a given key and direction.
 * String comparison is locale-aware; numbers use numeric subtraction.
 * Returns a new array and does not mutate the input.
 */
export function sortRows<T extends Record<string, unknown>>(rows: T[], sort: SortState): T[] {
  return [...rows].sort((a, b) => {
    const av = a[sort.key];
    const bv = b[sort.key];

    if (typeof av === 'number' && typeof bv === 'number') {
      return sort.direction === 'asc' ? av - bv : bv - av;
    }

    const as = String(av ?? '');
    const bs = String(bv ?? '');
    const cmp = as.localeCompare(bs, undefined, { numeric: true, sensitivity: 'base' });
    return sort.direction === 'asc' ? cmp : -cmp;
  });
}

/**
 * toggleSort returns the next SortState for a column header click.
 * Clicking the same column flips asc/desc; clicking a new column
 * defaults to ascending.
 */
export function toggleSort(current: SortState | null, key: string): SortState {
  if (current?.key === key) {
    return { key, direction: current.direction === 'asc' ? 'desc' : 'asc' };
  }
  return { key, direction: 'asc' };
}

/**
 * paginateRows slices a row array to the given page (1-indexed) and
 * page size. Returns an empty array when page is out of range.
 */
export function paginateRows<T>(rows: T[], page: number, pageSize: number): T[] {
  if (pageSize <= 0 || page < 1) return [];
  const start = (page - 1) * pageSize;
  return rows.slice(start, start + pageSize);
}

/**
 * pageCount returns the total number of pages for a given row count
 * and page size. Returns 0 when rowCount is 0.
 */
export function pageCount(rowCount: number, pageSize: number): number {
  if (rowCount <= 0 || pageSize <= 0) return 0;
  return Math.ceil(rowCount / pageSize);
}

/**
 * pageRange returns the human-readable label describing the visible
 * rows, e.g. "1-25 of 100".
 */
export function pageRange(rowCount: number, page: number, pageSize: number): string {
  if (rowCount === 0) return '0 results';
  const start = (page - 1) * pageSize + 1;
  const end = Math.min(page * pageSize, rowCount);
  return `${start}-${end} of ${rowCount}`;
}
