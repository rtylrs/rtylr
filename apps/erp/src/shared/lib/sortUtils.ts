/**
 * Sorting and comparison helpers for the ERP application.
 * All functions are pure with no side-effects.
 */

export type SortDir = 'asc' | 'desc';

export interface SortState<T extends string = string> {
  column: T | null;
  dir: SortDir;
}

/**
 * Compares two primitive values. Nulls sort after non-nulls.
 * Strings are compared with localeCompare.
 */
export function compareValues(a: unknown, b: unknown): number {
  if (a === null || a === undefined) return 1;
  if (b === null || b === undefined) return -1;
  if (typeof a === 'string' && typeof b === 'string') return a.localeCompare(b);
  if (a < b) return -1;
  if (a > b) return 1;
  return 0;
}

/**
 * Returns a comparator for sorting objects by a single key.
 */
export function compareBy<T extends object>(
  key: keyof T,
  dir: SortDir = 'asc'
): (a: T, b: T) => number {
  return (a, b) => {
    const result = compareValues(a[key], b[key]);
    return dir === 'asc' ? result : -result;
  };
}

export interface SortSpec<T extends object> {
  key: keyof T;
  dir?: SortDir;
}

/**
 * Returns a comparator that sorts by multiple keys in order.
 */
export function compareByMultiple<T extends object>(specs: SortSpec<T>[]): (a: T, b: T) => number {
  return (a, b) => {
    for (const { key, dir = 'asc' } of specs) {
      const result = compareValues(a[key], b[key]);
      if (result !== 0) return dir === 'asc' ? result : -result;
    }
    return 0;
  };
}

/**
 * Returns a sorted copy of the array without mutating the original.
 */
export function sortWith<T>(items: T[], compareFn: (a: T, b: T) => number): T[] {
  return [...items].sort(compareFn);
}

/** Toggles a sort direction. */
export function toggleSortDir(dir: SortDir): SortDir {
  return dir === 'asc' ? 'desc' : 'asc';
}

/**
 * Computes the next sort state when a column header is clicked.
 * Clicking a new column sets it to ascending; clicking the same column toggles direction.
 */
export function nextSortState<T extends string>(column: T, current: SortState<T>): SortState<T> {
  if (current.column === column) {
    return { column, dir: toggleSortDir(current.dir) };
  }
  return { column, dir: 'asc' };
}
