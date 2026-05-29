/** Lightweight cryptography helpers for the auth app. */

/**
 * Encodes a string to base64.
 */
export function toBase64(value: string): string {
  if (typeof btoa !== 'undefined') return btoa(value);
  return Buffer.from(value, 'utf8').toString('base64');
}

/**
 * Decodes a base64 string.
 */
export function fromBase64(value: string): string {
  if (typeof atob !== 'undefined') return atob(value);
  return Buffer.from(value, 'base64').toString('utf8');
}

/**
 * Encodes a string to URL-safe base64 (replaces +, /, and trims =).
 */
export function toBase64Url(value: string): string {
  return toBase64(value).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
}

/**
 * Decodes a URL-safe base64 string.
 */
export function fromBase64Url(value: string): string {
  const padded = value.replace(/-/g, '+').replace(/_/g, '/');
  const padding = (4 - (padded.length % 4)) % 4;
  return fromBase64(padded + '='.repeat(padding));
}

/**
 * Returns a simple non-cryptographic hash of the string (djb2 algorithm).
 * Suitable for checksums and cache keys, NOT for security purposes.
 */
export function hashString(value: string): number {
  let hash = 5381;
  for (let i = 0; i < value.length; i++) {
    hash = ((hash << 5) + hash) ^ value.charCodeAt(i);
    hash = hash >>> 0; // convert to unsigned 32-bit integer
  }
  return hash;
}
