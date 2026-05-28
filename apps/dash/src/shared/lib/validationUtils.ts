/**
 * Form validation utility helpers for the Dash app.
 */

/** Check if a string is a valid email address. */
export function isValidEmail(email: string): boolean {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email.trim());
}

/** Check if a string is a valid URL. */
export function isValidUrl(url: string): boolean {
  try {
    const u = new URL(url);
    return u.protocol === 'http:' || u.protocol === 'https:';
  } catch {
    return false;
  }
}

/** Check if a phone number contains only digits, spaces, and common separators. */
export function isValidPhone(phone: string): boolean {
  return /^[\d\s+\-().]{7,20}$/.test(phone.trim());
}

/** Return an error message if the string is required but empty. */
export function requireField(value: string, fieldName = 'Field'): string | null {
  return value.trim() ? null : `${fieldName} is required`;
}

/** Return an error message if the string is shorter than `min` characters. */
export function minLength(value: string, min: number, fieldName = 'Field'): string | null {
  return value.trim().length >= min ? null : `${fieldName} must be at least ${min} characters`;
}

/** Return an error message if the string is longer than `max` characters. */
export function maxLength(value: string, max: number, fieldName = 'Field'): string | null {
  return value.length <= max ? null : `${fieldName} must be at most ${max} characters`;
}

/** Return an error message if the number is outside [min, max]. */
export function numberRange(
  value: number,
  min: number,
  max: number,
  fieldName = 'Value'
): string | null {
  if (value < min) return `${fieldName} must be at least ${min}`;
  if (value > max) return `${fieldName} must be at most ${max}`;
  return null;
}

/** Collect all non-null validation errors into an array. */
export function collectErrors(...validators: Array<string | null>): string[] {
  return validators.filter((e): e is string => e !== null);
}

/** Return true when all validators pass (no errors). */
export function isValid(...validators: Array<string | null>): boolean {
  return validators.every((v) => v === null);
}
