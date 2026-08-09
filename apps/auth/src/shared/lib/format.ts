/**
 * Lightweight formatting helpers for the auth app.
 * All functions are pure and have no side-effects.
 */

/**
 * Truncates a string to maxLength characters, appending an ellipsis when cut.
 */
export function truncate(value: string, maxLength: number): string {
  if (value.length <= maxLength) return value;
  return `${value.slice(0, maxLength)}...`;
}

/**
 * Capitalises the first character of a string.
 */
export function capitalise(value: string): string {
  if (!value) return '';
  return value.charAt(0).toUpperCase() + value.slice(1);
}

/**
 * Derives a display name from an email address.
 * Uses the local-part (before @) with dots/underscores replaced by spaces,
 * then title-cases each word.
 *
 * @example nameFromEmail('john.doe@example.com') // 'John Doe'
 */
export function nameFromEmail(email: string): string {
  const local = email.split('@')[0] ?? '';
  return local
    .replace(/[._-]+/g, ' ')
    .split(' ')
    .map((word) => capitalise(word))
    .join(' ');
}

/**
 * Masks an email address so only the first character and domain are visible.
 *
 * @example maskEmail('john.doe@example.com') // 'j***@example.com'
 */
export function maskEmail(email: string): string {
  const atIdx = email.indexOf('@');
  if (atIdx <= 0) return email;
  return `${email.charAt(0)}***${email.slice(atIdx)}`;
}

/**
 * Returns true when the string is a valid URL (http or https).
 */
export function isValidUrl(value: string): boolean {
  try {
    const url = new URL(value);
    return url.protocol === 'http:' || url.protocol === 'https:';
  } catch {
    return false;
  }
}

/**
 * Strips leading/trailing whitespace and collapses internal runs of
 * whitespace to a single space.
 */
export function normaliseWhitespace(value: string): string {
  return value.trim().replace(/\s+/g, ' ');
}
