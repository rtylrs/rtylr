/**
 * Form field validation helpers for the landing app.
 * Pure functions with no external dependencies.
 */

/** Returns true when the string is a valid email address. */
export function isValidEmail(value: string): boolean {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value.trim());
}

/** Returns true when the string is a valid US or international phone number. */
export function isValidPhone(value: string): boolean {
  const digits = value.replace(/\D/g, '');
  return digits.length >= 7 && digits.length <= 15;
}

/** Returns true when the string is a valid URL. */
export function isValidUrl(value: string): boolean {
  try {
    const url = new URL(value);
    return url.protocol === 'http:' || url.protocol === 'https:';
  } catch {
    return false;
  }
}

/** Returns true when the string is non-empty after trimming. */
export function isRequired(value: string): boolean {
  return value.trim().length > 0;
}

/** Returns true when the string length is within the given bounds (inclusive). */
export function isLengthBetween(value: string, min: number, max: number): boolean {
  const len = value.trim().length;
  return len >= min && len <= max;
}

/** Returns true when the string contains only alphabetic characters and spaces. */
export function isAlpha(value: string): boolean {
  return /^[A-Za-z\s]+$/.test(value);
}

/** Returns true when the string contains only alphanumeric characters. */
export function isAlphanumeric(value: string): boolean {
  return /^[A-Za-z0-9]+$/.test(value);
}

/** Returns true when the value is a finite number within an optional range. */
export function isNumericInRange(value: number, min?: number, max?: number): boolean {
  if (!isFinite(value)) return false;
  if (min !== undefined && value < min) return false;
  if (max !== undefined && value > max) return false;
  return true;
}

/** Returns true when both strings match (useful for confirm-password fields). */
export function isMatch(a: string, b: string): boolean {
  return a === b;
}

/** Returns true when the string looks like a valid zip/postal code (US 5-digit or CA). */
export function isPostalCode(value: string): boolean {
  return (
    /^\d{5}(-\d{4})?$/.test(value.trim()) ||
    /^[A-Za-z]\d[A-Za-z][ -]?\d[A-Za-z]\d$/.test(value.trim())
  );
}

/** Validates that a password meets minimum strength requirements. */
export function isStrongPassword(value: string): boolean {
  if (value.length < 8) return false;
  const hasUpper = /[A-Z]/.test(value);
  const hasLower = /[a-z]/.test(value);
  const hasDigit = /\d/.test(value);
  return hasUpper && hasLower && hasDigit;
}

/**
 * Returns a list of validation errors for a password.
 * Empty array means the password passes all checks.
 */
export function passwordErrors(value: string): string[] {
  const errors: string[] = [];
  if (value.length < 8) errors.push('At least 8 characters required');
  if (!/[A-Z]/.test(value)) errors.push('At least one uppercase letter required');
  if (!/[a-z]/.test(value)) errors.push('At least one lowercase letter required');
  if (!/\d/.test(value)) errors.push('At least one number required');
  return errors;
}
