/**
 * Validation helper functions used across ERP forms and inputs.
 */

/** Return true when a string is a valid email address. */
export function isValidEmail(email: string): boolean {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email.trim());
}

/** Return true when a phone number contains at least 7 digits. */
export function isValidPhone(phone: string): boolean {
  const digits = phone.replace(/\D/g, '');
  return digits.length >= 7;
}

/** Return true when a value is a positive number greater than zero. */
export function isPositiveNumber(value: number): boolean {
  return isFinite(value) && value > 0;
}

/** Return true when a value is a non-negative number (zero or greater). */
export function isNonNegative(value: number): boolean {
  return isFinite(value) && value >= 0;
}

/** Return true when a string is non-empty after trimming. */
export function isNonEmpty(value: string): boolean {
  return value.trim().length > 0;
}

/** Return true when a value falls within [min, max] inclusive. */
export function isInRange(value: number, min: number, max: number): boolean {
  return value >= min && value <= max;
}

/** Return a validation error string or null for a required text field. */
export function requiredFieldError(value: string, fieldName: string): string | null {
  if (!isNonEmpty(value)) return `${fieldName} is required`;
  return null;
}

/** Return a validation error string or null for a positive number field. */
export function positiveNumberError(value: number, fieldName: string): string | null {
  if (!isPositiveNumber(value)) return `${fieldName} must be greater than 0`;
  return null;
}

/** Return true when a string matches the pattern of a URL (http/https). */
export function isValidUrl(url: string): boolean {
  try {
    const u = new URL(url);
    return u.protocol === 'http:' || u.protocol === 'https:';
  } catch {
    return false;
  }
}
