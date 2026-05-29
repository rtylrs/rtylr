/**
 * Generic search and filter utilities for the Menu app.
 */

/** Return true if a string haystack contains the needle (case-insensitive). */
export function containsQuery(haystack: string, needle: string): boolean {
  if (!needle.trim()) return true;
  return haystack.toLowerCase().includes(needle.trim().toLowerCase());
}

/** Split a query string into individual tokens. */
export function tokenise(query: string): string[] {
  return query
    .trim()
    .toLowerCase()
    .split(/\s+/)
    .filter((t) => t.length > 0);
}

/** Return true if all tokens appear in the haystack. */
export function matchesAllTokens(haystack: string, query: string): boolean {
  const tokens = tokenise(query);
  if (tokens.length === 0) return true;
  const lower = haystack.toLowerCase();
  return tokens.every((token) => lower.includes(token));
}

/** Highlight occurrences of a query within text by wrapping them with markers. */
export function highlight(text: string, query: string, open = '<mark>', close = '</mark>'): string {
  if (!query.trim()) return text;
  const escaped = query.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
  return text.replace(new RegExp(escaped, 'gi'), (match) => `${open}${match}${close}`);
}

/** Filter an array of items by running a query against a string field. */
export function filterByField<T>(items: T[], getField: (item: T) => string, query: string): T[] {
  if (!query.trim()) return items;
  return items.filter((item) => containsQuery(getField(item), query));
}

/** Sort items so that those whose names start with the query appear first. */
export function sortByRelevance<T>(items: T[], getField: (item: T) => string, query: string): T[] {
  if (!query.trim()) return items;
  const q = query.trim().toLowerCase();
  return [...items].sort((a, b) => {
    const aStarts = getField(a).toLowerCase().startsWith(q);
    const bStarts = getField(b).toLowerCase().startsWith(q);
    if (aStarts && !bStarts) return -1;
    if (!aStarts && bStarts) return 1;
    return 0;
  });
}
