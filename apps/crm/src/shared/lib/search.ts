// Client-side search/filter helpers for CRM list views.

/**
 * Returns true when the haystack string contains the needle,
 * using case-insensitive matching after trimming both values.
 */
export function matchesSearch(haystack: string, needle: string): boolean {
  return haystack.toLowerCase().includes(needle.trim().toLowerCase());
}

/**
 * Filter a list of objects by a search string against multiple fields.
 * Returns all items when `query` is empty.
 */
export function filterByQuery<T extends Record<string, unknown>>(
  items: T[],
  query: string,
  fields: (keyof T)[]
): T[] {
  const q = query.trim().toLowerCase();
  if (!q) return items;

  return items.filter((item) =>
    fields.some((field) => {
      const value = item[field];
      return typeof value === 'string' && value.toLowerCase().includes(q);
    })
  );
}

/**
 * Return the subset of `items` where the value at `key` equals `value`.
 * Returns all items when `value` is undefined or null.
 */
export function filterByValue<T>(items: T[], key: keyof T, value: T[keyof T] | undefined): T[] {
  if (value == null) return items;
  return items.filter((item) => item[key] === value);
}

/**
 * Return the subset of `items` where the array at `key` contains `tag`.
 */
export function filterByTag<T>(items: T[], key: keyof T, tag: string): T[] {
  if (!tag) return items;
  return items.filter((item) => {
    const val = item[key];
    return Array.isArray(val) && val.includes(tag);
  });
}

/**
 * Collect all unique string values across items for a given array field.
 * Useful for building tag/label filter dropdowns.
 */
export function collectTags<T>(items: T[], key: keyof T): string[] {
  const set = new Set<string>();
  for (const item of items) {
    const val = item[key];
    if (Array.isArray(val)) {
      for (const tag of val) {
        if (typeof tag === 'string') set.add(tag);
      }
    }
  }
  return Array.from(set).sort();
}
