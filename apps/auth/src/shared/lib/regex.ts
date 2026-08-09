/** Regex pattern constants and matcher helpers for the auth app. */

export const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

export const STRONG_PASSWORD_RE = /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[^a-zA-Z\d]).{8,}$/;

export const PHONE_RE = /^\+?[\d\s\-().]{7,20}$/;

export const UUID_V4_RE = /^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;

export const URL_RE = /^https?:\/\/[\w\-]+(\.[\w\-]+)+([\w\-.,@?^=%&:/~+#]*[\w\-@?^=%&/~+#])?$/;

export const SLUG_RE = /^[a-z0-9]+(?:-[a-z0-9]+)*$/;

export const HEX_COLOR_RE = /^#([0-9a-f]{3}|[0-9a-f]{6})$/i;

/**
 * Returns true if value matches the given regex pattern.
 */
export function matches(value: string, pattern: RegExp): boolean {
  return pattern.test(value);
}
