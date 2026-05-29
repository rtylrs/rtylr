/**
 * Truncates a string to the given max length, appending an ellipsis when
 * the string is longer than the limit. Does not split in the middle of a word.
 */
export function truncate(text: string, maxLength: number, ellipsis = '...'): string {
  if (text.length <= maxLength) return text;
  const trimmed = text.slice(0, maxLength - ellipsis.length).trimEnd();
  return trimmed + ellipsis;
}

/**
 * Truncates a string at exactly maxLength characters, appending an ellipsis.
 * Unlike truncate(), this does not respect word boundaries.
 */
export function truncateHard(text: string, maxLength: number, ellipsis = '...'): string {
  if (text.length <= maxLength) return text;
  return text.slice(0, maxLength - ellipsis.length) + ellipsis;
}

/**
 * Returns true when the string exceeds the given max length.
 */
export function isTruncated(text: string, maxLength: number): boolean {
  return text.length > maxLength;
}
