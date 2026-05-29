/**
 * Search and filter utility helpers for the Dash app.
 * All functions are pure with no side-effects.
 */

/**
 * Return true when haystack contains needle (case-insensitive).
 */
export function matchesText(haystack: string, needle: string): boolean {
  return haystack.toLowerCase().includes(needle.toLowerCase());
}

/**
 * Filter an array of objects by a text query across specified fields.
 * Returns all items when the query is empty or whitespace-only.
 */
export function filterByText<T extends Record<string, unknown>>(
  items: T[],
  query: string,
  fields: (keyof T)[]
): T[] {
  const trimmed = query.trim();
  if (!trimmed) return items;
  return items.filter((item) =>
    fields.some((field) => {
      const value = item[field];
      return typeof value === 'string' && matchesText(value, trimmed);
    })
  );
}

/**
 * Return unique non-empty string values from the specified field.
 * Preserves insertion order.
 */
export function uniqueFieldValues<T extends Record<string, unknown>>(
  items: T[],
  field: keyof T
): string[] {
  const seen = new Set<string>();
  const result: string[] = [];
  for (const item of items) {
    const val = item[field];
    if (typeof val === 'string' && val.trim() && !seen.has(val)) {
      seen.add(val);
      result.push(val);
    }
  }
  return result;
}

/**
 * Highlight occurrences of a query in text by wrapping them.
 * Returns the original text when query is empty.
 */
export function highlightMatches(
  text: string,
  query: string,
  wrapper = (match: string) => `<mark>${match}</mark>`
): string {
  if (!query.trim()) return text;
  const escaped = query.trim().replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
  const regex = new RegExp(`(${escaped})`, 'gi');
  return text.replace(regex, (match) => wrapper(match));
}

/**
 * Split a search query into individual terms by whitespace.
 * Filters out empty strings.
 */
export function parseSearchTerms(query: string): string[] {
  return query.trim().split(/\s+/).filter(Boolean);
}

/**
 * Return true when text matches ALL terms in the query (AND search).
 */
export function matchesAllTerms(text: string, query: string): boolean {
  const terms = parseSearchTerms(query);
  if (terms.length === 0) return true;
  return terms.every((term) => matchesText(text, term));
}
