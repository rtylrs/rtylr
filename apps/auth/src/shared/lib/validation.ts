/**
 * Small shared validation helpers used by the auth screens.
 * Keep these dependency-free so they can run anywhere.
 */

export const isEmail = (value: string): boolean => {
  if (!value) return false;
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value.trim());
};

export const isStrongPassword = (value: string): boolean => {
  if (!value || value.length < 8) return false;
  const hasUpper = /[A-Z]/.test(value);
  const hasLower = /[a-z]/.test(value);
  const hasDigit = /\d/.test(value);
  return hasUpper && hasLower && hasDigit;
};

export const passwordsMatch = (a: string, b: string): boolean => !!a && a === b;

export const isPhoneNumber = (value: string): boolean => {
  if (!value) return false;
  return /^\+?[0-9\s-]{6,}$/.test(value.trim());
};

export const isNotEmpty = (value: string): boolean => value.trim().length > 0;

export const getEmailError = (value: string): string | undefined => {
  if (!value) return undefined;
  return isEmail(value) ? undefined : 'Please enter a valid email address';
};

export const getPasswordMatchError = (a: string, b: string): string | undefined => {
  if (!b) return undefined;
  return passwordsMatch(a, b) ? undefined : 'Passwords do not match';
};
