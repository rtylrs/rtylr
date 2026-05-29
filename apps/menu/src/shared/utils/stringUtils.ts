/**
 * String utility helpers for the menu app.
 */

/**
 * Capitalises the first character of a string.
 */
export function capitalise(text: string): string {
  if (!text) return text;
  return text.charAt(0).toUpperCase() + text.slice(1);
}

/**
 * Converts a camelCase or PascalCase string to kebab-case.
 */
export function toKebabCase(text: string): string {
  return text
    .replace(/([a-z])([A-Z])/g, '$1-$2')
    .replace(/\s+/g, '-')
    .toLowerCase();
}

/**
 * Converts a kebab-case or snake_case string to Title Case.
 */
export function toTitleCase(text: string): string {
  return text
    .replace(/[-_]/g, ' ')
    .split(' ')
    .map((word) => capitalise(word))
    .join(' ');
}

/**
 * Returns true if the string contains only whitespace or is empty.
 */
export function isBlank(text: string): boolean {
  return text.trim().length === 0;
}

/**
 * Strips all HTML tags from a string.
 */
export function stripHtml(text: string): string {
  return text.replace(/<[^>]*>/g, '');
}

/**
 * Counts the number of words in a string.
 */
export function wordCount(text: string): number {
  const trimmed = text.trim();
  if (!trimmed) return 0;
  return trimmed.split(/\s+/).length;
}

/**
 * Repeats a string n times.
 */
export function repeat(text: string, times: number): string {
  if (times <= 0) return '';
  return text.repeat(times);
}

/**
 * Pads a string on the left to the specified length with the given fill character.
 */
export function padLeft(text: string, length: number, fill = ' '): string {
  return text.padStart(length, fill);
}
