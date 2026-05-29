// ERP search and pagination utilities.

/**
 * Returns true when haystack contains needle (case-insensitive).
 */
export function matchesSearch(haystack: string, needle: string): boolean {
  return haystack.toLowerCase().includes(needle.toLowerCase());
}

/**
 * Filters an array of objects by a search query across the given field keys.
 * Returns all items when query is empty or only whitespace.
 */
export function filterBySearch<T extends Record<string, unknown>>(
  items: T[],
  query: string,
  fields: (keyof T)[]
): T[] {
  const q = query.trim();
  if (!q) return items;
  return items.filter((item) =>
    fields.some((field) => {
      const val = item[field];
      if (typeof val !== 'string') return false;
      return matchesSearch(val, q);
    })
  );
}

export interface PaginationResult<T> {
  items: T[];
  page: number;
  pageSize: number;
  totalItems: number;
  totalPages: number;
}

/**
 * Paginates an array, clamping page to valid range.
 */
export function paginate<T>(items: T[], page: number, pageSize: number): PaginationResult<T> {
  const totalItems = items.length;
  const totalPages = Math.max(1, Math.ceil(totalItems / pageSize));
  const safePage = Math.max(1, Math.min(page, totalPages));
  const start = (safePage - 1) * pageSize;
  return {
    items: items.slice(start, start + pageSize),
    page: safePage,
    pageSize,
    totalItems,
    totalPages,
  };
}

/**
 * Removes duplicates from a string array, preserving insertion order.
 */
export function uniqueStrings(values: string[]): string[] {
  return [...new Set(values)];
}

/**
 * Returns unique values for the given key from an array of objects,
 * preserving insertion order.
 */
export function uniqueValues<T extends Record<string, unknown>>(
  items: T[],
  key: keyof T
): T[keyof T][] {
  const seen = new Set<unknown>();
  const result: T[keyof T][] = [];
  for (const item of items) {
    const val = item[key];
    if (!seen.has(val)) {
      seen.add(val);
      result.push(val);
    }
  }
  return result;
}
