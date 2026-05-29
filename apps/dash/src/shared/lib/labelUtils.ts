/**
 * Text labeling and formatting helpers for the Dash app.
 * All functions are pure with no side-effects.
 */

/** Returns the singular or plural form based on count. */
export function pluralize(count: number, singular: string, plural?: string): string {
  return count === 1 ? singular : (plural ?? `${singular}s`);
}

/** Returns a count with the appropriate singular/plural label. */
export function countLabel(count: number, singular: string, plural?: string): string {
  return `${count} ${pluralize(count, singular, plural)}`;
}

/** Truncates a string to maxLen, appending ellipsis if truncated. */
export function truncate(value: string, maxLen: number, ellipsis: string = '...'): string {
  if (value.length <= maxLen) return value;
  return value.slice(0, maxLen - ellipsis.length) + ellipsis;
}

/** Capitalizes the first character of a string. */
export function capitalize(value: string): string {
  if (!value) return value;
  return value.charAt(0).toUpperCase() + value.slice(1);
}

/**
 * Converts a camelCase or snake_case key to a Title Case label.
 * e.g. 'firstName' -> 'First Name', 'first_name' -> 'First Name'
 */
export function toLabel(key: string): string {
  // Replace underscores with spaces
  const spaced = key.replace(/_/g, ' ');
  // Insert space before uppercase letters (camelCase)
  const split = spaced.replace(/([a-z])([A-Z])/g, '$1 $2');
  // Capitalize each word
  return split
    .split(' ')
    .map((w) => capitalize(w))
    .join(' ');
}

/** Returns the ordinal string for a number. e.g. 1 -> '1st', 11 -> '11th', 21 -> '21st' */
export function ordinal(n: number): string {
  const abs = Math.abs(n);
  const mod100 = abs % 100;
  const mod10 = abs % 10;
  if (mod100 >= 11 && mod100 <= 13) return `${n}th`;
  if (mod10 === 1) return `${n}st`;
  if (mod10 === 2) return `${n}nd`;
  if (mod10 === 3) return `${n}rd`;
  return `${n}th`;
}

/**
 * Joins a list of strings with commas and a conjunction before the last item.
 * e.g. ['a', 'b', 'c'] -> 'a, b, and c'
 */
export function joinList(items: string[], conjunction: string = 'and'): string {
  if (items.length === 0) return '';
  if (items.length === 1) return items[0];
  if (items.length === 2) return `${items[0]} ${conjunction} ${items[1]}`;
  return `${items.slice(0, -1).join(', ')}, ${conjunction} ${items[items.length - 1]}`;
}

/** Returns true if the string is null, undefined, or whitespace-only. */
export function isBlank(value: string | null | undefined): boolean {
  return value == null || value.trim().length === 0;
}
