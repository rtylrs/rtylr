/** Capitalizes the first character of a string. */
export function capitalize(s: string): string {
  if (!s) return s;
  return s.charAt(0).toUpperCase() + s.slice(1);
}

/** Converts a snake_case string to Title Case. */
export function snakeToTitle(s: string): string {
  return s
    .split('_')
    .map((word) => capitalize(word))
    .join(' ');
}

/** Converts a camelCase string to snake_case. */
export function camelToSnake(s: string): string {
  return s.replace(/([A-Z])/g, (m) => `_${m.toLowerCase()}`);
}

/** Converts a snake_case string to camelCase. */
export function snakeToCamel(s: string): string {
  return s.replace(/_([a-z])/g, (_, c) => c.toUpperCase());
}

/** Truncates a string to the given max length, adding ellipsis if needed. */
export function truncate(s: string, maxLength: number, ellipsis = '...'): string {
  if (s.length <= maxLength) return s;
  return s.slice(0, maxLength - ellipsis.length) + ellipsis;
}

/** Returns true if the string is blank (null, undefined, or whitespace only). */
export function isBlank(s: string | null | undefined): boolean {
  return !s || s.trim().length === 0;
}

/** Returns a slug (URL-safe lowercase string) from the given string. */
export function toSlug(s: string): string {
  return s
    .toLowerCase()
    .trim()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '');
}

/** Returns initials from a full name (up to 2 characters). */
export function initials(name: string): string {
  return name
    .trim()
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((word) => word[0].toUpperCase())
    .join('');
}

/** Pluralizes a word based on count (naive English rules). */
export function pluralize(word: string, count: number, plural?: string): string {
  if (count === 1) return word;
  return plural ?? `${word}s`;
}

/** Returns the count and noun together, e.g. "3 tasks". */
export function countLabel(count: number, singular: string, plural?: string): string {
  return `${count} ${pluralize(singular, count, plural)}`;
}

/**
 * flowApi substitutes '-' for a description omitted at write time, because
 * several contracts mark the field @required. That sentinel is a wire-format
 * workaround only - never render it back to the user as if it were real
 * content. Use this wherever a description read from the API is displayed.
 */
export function displayDescription(description: string | undefined | null): string {
  if (!description || description === '-') return '';
  return description;
}
