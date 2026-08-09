/**
 * Derives initials from a display name.
 * Returns at most two characters, uppercased.
 */
export function getInitials(name: string): string {
  const parts = name.trim().split(/\s+/).filter(Boolean);
  if (parts.length === 0) return '';
  if (parts.length === 1) return parts[0][0].toUpperCase();
  return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
}

/**
 * Converts a string to a URL-safe slug.
 */
export function toSlug(value: string): string {
  return value
    .trim()
    .toLowerCase()
    .replace(/[\s_]+/g, '-')
    .replace(/[^a-z0-9-]/g, '')
    .replace(/-+/g, '-')
    .replace(/^-|-$/g, '');
}

/**
 * Returns true when the string contains only ASCII characters.
 */
export function isAscii(value: string): boolean {
  return /^[\x00-\x7F]*$/.test(value);
}

/**
 * Pads a string on the right to the given length.
 */
export function padRight(value: string, length: number, fill = ' '): string {
  if (value.length >= length) return value;
  return value + fill.repeat(length - value.length);
}

/**
 * Counts the occurrences of a substring within a string.
 */
export function countOccurrences(haystack: string, needle: string): number {
  if (!needle) return 0;
  let count = 0;
  let pos = 0;
  while ((pos = haystack.indexOf(needle, pos)) !== -1) {
    count++;
    pos += needle.length;
  }
  return count;
}

/**
 * Converts a camelCase or PascalCase identifier to a human-readable label.
 */
export function camelToLabel(value: string): string {
  return value
    .replace(/([A-Z]+)([A-Z][a-z])/g, '$1 $2')
    .replace(/([a-z\d])([A-Z])/g, '$1 $2')
    .replace(/^./, (s) => s.toUpperCase())
    .trim();
}
