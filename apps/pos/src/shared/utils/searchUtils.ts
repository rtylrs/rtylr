// Pure helpers for client-side search and filtering.

/** Normalize a search query: trim and lowercase. */
export function normalizeQuery(query: string): string {
  return query.trim().toLowerCase();
}

/** Returns true if a text field matches a query string (case-insensitive). */
export function matchesQuery(text: string, query: string): boolean {
  if (!query) return true;
  return text.toLowerCase().includes(query.toLowerCase());
}

/** Filter an array of objects by a query against multiple string fields. */
export function filterByQuery<T extends Record<string, unknown>>(
  items: T[],
  query: string,
  fields: (keyof T)[],
): T[] {
  const normalized = normalizeQuery(query);
  if (!normalized) return items;
  return items.filter((item) =>
    fields.some((field) => {
      const value = item[field];
      if (typeof value !== 'string') return false;
      return value.toLowerCase().includes(normalized);
    }),
  );
}

/** Sort an array of objects by a string field, ascending by default. */
export function sortByField<T extends Record<string, unknown>>(
  items: T[],
  field: keyof T,
  direction: 'asc' | 'desc' = 'asc',
): T[] {
  return [...items].sort((a, b) => {
    const aVal = String(a[field] ?? '');
    const bVal = String(b[field] ?? '');
    const cmp = aVal.localeCompare(bVal);
    return direction === 'asc' ? cmp : -cmp;
  });
}

/** Deduplicate an array by a key function, keeping the first occurrence. */
export function dedupeBy<T>(items: T[], keyFn: (item: T) => string): T[] {
  const seen = new Set<string>();
  return items.filter((item) => {
    const key = keyFn(item);
    if (seen.has(key)) return false;
    seen.add(key);
    return true;
  });
}
