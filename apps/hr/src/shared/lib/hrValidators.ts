/**
 * HR-specific validation helpers.
 * These are pure functions that return booleans or error message strings.
 */

/**
 * Returns true if the given string is a syntactically valid email address.
 */
export function isValidEmail(email: string): boolean {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email.trim());
}

/**
 * Returns true if the hire date is not in the future.
 */
export function isValidHireDate(hireDate: string): boolean {
  const d = new Date(hireDate);
  return !isNaN(d.getTime()) && d <= new Date();
}

/**
 * Returns true if the probation end date is after the hire date.
 */
export function isProbationEndAfterHire(hireDate: string, probationEnd: string): boolean {
  const hire = new Date(hireDate);
  const end = new Date(probationEnd);
  return !isNaN(hire.getTime()) && !isNaN(end.getTime()) && end > hire;
}

/**
 * Returns true if the leave request end date is on or after the start date.
 */
export function isValidLeaveRange(startDate: string, endDate: string): boolean {
  const start = new Date(startDate);
  const end = new Date(endDate);
  return !isNaN(start.getTime()) && !isNaN(end.getTime()) && end >= start;
}

/**
 * Returns true if the given hourly rate is a positive finite number.
 */
export function isValidHourlyRate(rate: number): boolean {
  return Number.isFinite(rate) && rate > 0;
}

/**
 * Returns true if the given monthly salary is a positive finite number.
 */
export function isValidMonthlySalary(salary: number): boolean {
  return Number.isFinite(salary) && salary > 0;
}

/**
 * Returns true if the given currency code is a 3-letter ISO 4217 string.
 */
export function isValidCurrencyCode(code: string): boolean {
  return /^[A-Z]{3}$/.test(code.trim());
}

/**
 * Returns true if the given PIN is exactly 4 digits.
 */
export function isValidPin(pin: string): boolean {
  return /^\d{4}$/.test(pin);
}

/**
 * Returns true if name is non-empty after trimming.
 */
export function isValidName(name: string): boolean {
  return name.trim().length > 0;
}

/**
 * Returns a human-readable error message for a leave range, or null if valid.
 */
export function validateLeaveRange(startDate: string, endDate: string): string | null {
  if (!startDate || !endDate) return 'Start and end dates are required';
  if (!isValidLeaveRange(startDate, endDate)) return 'End date must be on or after start date';
  return null;
}
