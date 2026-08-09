/**
 * Input validation helpers for the Menu app.
 * All functions are pure and return a boolean.
 */

/** Return true if the string is a valid email address. */
export function isValidEmail(email: string): boolean {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email.trim());
}

/** Return true if the string is not empty after trimming whitespace. */
export function isNonEmpty(value: string): boolean {
  return value.trim().length > 0;
}

/** Return true if the value is within [min, max] (inclusive). */
export function isInRange(value: number, min: number, max: number): boolean {
  return value >= min && value <= max;
}

/** Return true if the string contains only digits. */
export function isDigitsOnly(value: string): boolean {
  return /^\d+$/.test(value);
}

/** Return true if the string length is within [min, max] (inclusive). */
export function isLengthBetween(value: string, min: number, max: number): boolean {
  return value.length >= min && value.length <= max;
}

/** Return true if the value is a positive integer. */
export function isPositiveInt(value: number): boolean {
  return Number.isInteger(value) && value > 0;
}

/** Return true if the string is a valid URL (http or https only). */
export function isValidUrl(value: string): boolean {
  try {
    const url = new URL(value);
    return url.protocol === 'http:' || url.protocol === 'https:';
  } catch {
    return false;
  }
}

/** Return true if the quantity is a non-negative integer. */
export function isValidQuantity(value: number): boolean {
  return Number.isInteger(value) && value >= 0;
}
