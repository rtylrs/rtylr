/** Sort helpers for the docs app. */

export interface Sortable {
  title?: string;
  date?: string | Date;
  order?: number;
}

/**
 * Sorts items alphabetically by title (case-insensitive).
 */
export function sortByTitle<T extends Sortable>(items: T[]): T[] {
  return [...items].sort((a, b) => (a.title ?? '').localeCompare(b.title ?? ''));
}

/**
 * Sorts items by date descending (newest first).
 */
export function sortByDate<T extends Sortable>(items: T[]): T[] {
  return [...items].sort((a, b) => {
    const da = a.date ? new Date(a.date).getTime() : 0;
    const db = b.date ? new Date(b.date).getTime() : 0;
    return db - da;
  });
}

/**
 * Sorts items by the `order` field ascending. Items without order go last.
 */
export function sortByOrder<T extends Sortable>(items: T[]): T[] {
  return [...items].sort((a, b) => {
    const oa = a.order ?? Infinity;
    const ob = b.order ?? Infinity;
    return oa - ob;
  });
}

/**
 * Reverses an array without mutating the original.
 */
export function reverse<T>(items: T[]): T[] {
  return [...items].reverse();
}
