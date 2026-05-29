/**
 * String utility helpers for the Dash app.
 * All functions are pure with no side-effects.
 */

/** Truncate a string to maxLen, appending ellipsis when cut. */
export function truncate(str: string, maxLen: number, ellipsis = '...'): string {
  if (str.length <= maxLen) return str;
  return str.slice(0, maxLen - ellipsis.length).trimEnd() + ellipsis;
}

/** Capitalize the first letter of a string. */
export function capitalize(str: string): string {
  if (!str) return str;
  return str.charAt(0).toUpperCase() + str.slice(1);
}

/** Convert camelCase to a human-readable label. e.g. "myFieldName" -> "My Field Name" */
export function camelToLabel(str: string): string {
  return str
    .replace(/([A-Z])/g, ' $1')
    .replace(/^./, (c) => c.toUpperCase())
    .trim();
}

/** Convert snake_case to Title Case. e.g. "my_field_name" -> "My Field Name" */
export function snakeToTitle(str: string): string {
  return str
    .split('_')
    .map((word) => capitalize(word))
    .join(' ');
}

/** Convert kebab-case to Title Case. e.g. "my-field-name" -> "My Field Name" */
export function kebabToTitle(str: string): string {
  return str
    .split('-')
    .map((word) => capitalize(word))
    .join(' ');
}

/** Left-pad a string to the given width with a fill character. */
export function padStart(str: string, width: number, fill = ' '): string {
  if (str.length >= width) return str;
  return fill.repeat(width - str.length) + str;
}

/** Remove all whitespace from a string. */
export function stripWhitespace(str: string): string {
  return str.replace(/\s+/g, '');
}

/** Count occurrences of a substring within a string. */
export function countOccurrences(str: string, substring: string): number {
  if (!substring) return 0;
  let count = 0;
  let pos = 0;
  while ((pos = str.indexOf(substring, pos)) !== -1) {
    count++;
    pos += substring.length;
  }
  return count;
}
