// String utilities for the insights app.
// Shared helpers for truncation, capitalization, and search matching
// used by search boxes, tooltips, and table renderers.

/**
 * truncate shortens a string to maxLength chars and appends ellipsis
 * when truncated. Returns the original string when it fits.
 */
export function truncate(value: string, maxLength: number, ellipsis = '...'): string {
  if (value.length <= maxLength) return value;
  return value.slice(0, maxLength - ellipsis.length) + ellipsis;
}

/**
 * capitalize returns the string with the first character uppercased
 * and the rest unchanged.
 */
export function capitalize(value: string): string {
  if (!value) return value;
  return value.charAt(0).toUpperCase() + value.slice(1);
}

/**
 * titleCase converts a string to Title Case by capitalizing each word.
 */
export function titleCase(value: string): string {
  return value
    .split(' ')
    .map((word) => (word ? capitalize(word) : word))
    .join(' ');
}

/**
 * camelToSnake converts a camelCase identifier to snake_case.
 * e.g. "totalRevenue" -> "total_revenue"
 */
export function camelToSnake(value: string): string {
  return value.replace(/([A-Z])/g, (match) => `_${match.toLowerCase()}`);
}

/**
 * snakeToCamel converts a snake_case identifier to camelCase.
 * e.g. "total_revenue" -> "totalRevenue"
 */
export function snakeToCamel(value: string): string {
  return value.replace(/_([a-z])/g, (_, char: string) => char.toUpperCase());
}

/**
 * highlightMatch wraps all occurrences of query within text with
 * <mark> tags for search result highlighting. Case-insensitive.
 * Returns the original text when query is empty.
 */
export function highlightMatch(text: string, query: string): string {
  if (!query.trim()) return text;
  const escaped = query.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
  return text.replace(new RegExp(escaped, 'gi'), (match) => `<mark>${match}</mark>`);
}

/**
 * initials returns up to 2 uppercase initials from a full name.
 * e.g. "Alice Smith" -> "AS", "Bob" -> "B"
 */
export function initials(name: string): string {
  const parts = name.trim().split(/\s+/).filter(Boolean);
  if (parts.length === 0) return '';
  if (parts.length === 1) return parts[0].charAt(0).toUpperCase();
  return (parts[0].charAt(0) + parts[parts.length - 1].charAt(0)).toUpperCase();
}
