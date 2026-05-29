/**
 * Table sort and filter utility helpers for the Dash app.
 * All functions are pure with no side-effects.
 */

export type SortDirection = 'asc' | 'desc';

export interface SortState<K extends string = string> {
  key: K;
  direction: SortDirection;
}

/** Toggle sort direction or set a new sort key with ascending direction. */
export function toggleSort<K extends string>(current: SortState<K> | null, key: K): SortState<K> {
  if (current?.key === key) {
    return { key, direction: current.direction === 'asc' ? 'desc' : 'asc' };
  }
  return { key, direction: 'asc' };
}

/** Sort an array of objects by a field key and direction. Does not mutate. */
export function sortRows<T extends Record<string, unknown>>(
  rows: T[],
  sort: SortState<Extract<keyof T, string>> | null
): T[] {
  if (!sort) return rows;
  const { key, direction } = sort;
  return [...rows].sort((a, b) => {
    const av = a[key];
    const bv = b[key];
    if (av === bv) return 0;
    const less = av == null || av < bv!;
    return direction === 'asc' ? (less ? -1 : 1) : less ? 1 : -1;
  });
}

/** Filter rows where any string field contains the query (case-insensitive). */
export function filterRows<T extends Record<string, unknown>>(
  rows: T[],
  query: string,
  fields?: (keyof T)[]
): T[] {
  const q = query.trim().toLowerCase();
  if (!q) return rows;
  const keys = fields ?? (Object.keys(rows[0] ?? {}) as (keyof T)[]);
  return rows.filter((row) =>
    keys.some((k) =>
      String(row[k] ?? '')
        .toLowerCase()
        .includes(q)
    )
  );
}

/** Return a slice of rows for the given 1-based page number and page size. */
export function pageRows<T>(rows: T[], page: number, pageSize: number): T[] {
  const start = (page - 1) * pageSize;
  return rows.slice(start, start + pageSize);
}

/** Return the total number of pages for the given row count and page size. */
export function totalPages(total: number, pageSize: number): number {
  if (pageSize <= 0) return 0;
  return Math.ceil(total / pageSize);
}

/** Return the visible page window centered around the current page. */
export function visiblePages(currentPage: number, total: number, windowSize = 5): number[] {
  if (total <= 0) return [];
  const half = Math.floor(windowSize / 2);
  let start = Math.max(1, currentPage - half);
  const end = Math.min(total, start + windowSize - 1);
  start = Math.max(1, end - windowSize + 1);
  const pages: number[] = [];
  for (let i = start; i <= end; i++) pages.push(i);
  return pages;
}

/** Return a label like "1-25 of 100" for the current page state. */
export function rowRangeLabel(page: number, pageSize: number, total: number): string {
  if (total === 0) return '0 results';
  const from = (page - 1) * pageSize + 1;
  const to = Math.min(page * pageSize, total);
  return `${from}-${to} of ${total}`;
}
