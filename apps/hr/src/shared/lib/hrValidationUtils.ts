/**
 * HR-specific validation helpers.
 * All functions are pure and have no side-effects.
 */
import { todayIso } from './dateUtils';

/** Return true if the string is a valid email address. */
export function isValidEmail(email: string): boolean {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email.trim());
}

/** Return true if the value looks like a phone number (7-15 digits, optional leading +). */
export function isValidPhone(phone: string): boolean {
  const digits = phone.replace(/[\s\-().+]/g, '');
  return /^\d{7,15}$/.test(digits);
}

/** Return true if the value is a numeric PIN of the specified length. */
export function isValidPin(pin: string, length = 4): boolean {
  return new RegExp(`^\\d{${length}}$`).test(pin);
}

/** Return true if the value is a valid person name (2+ chars, letters, hyphens, apostrophes). */
export function isValidName(name: string): boolean {
  return /^[A-Za-z\-' ]{2,}$/.test(name.trim()) && !/\d/.test(name);
}

/** Return true if the string is a valid YYYY-MM-DD date. */
export function isValidIsoDate(value: string): boolean {
  if (!/^\d{4}-\d{2}-\d{2}$/.test(value)) return false;
  const d = new Date(value);
  return !isNaN(d.getTime()) && d.toISOString().startsWith(value);
}

/** Return true if the hourly rate is within a reasonable range (0.01 - 10000). */
export function isValidHourlyRate(rate: number): boolean {
  return rate > 0 && rate <= 10000;
}

/** Return true if the annual salary is within a reasonable range (1 - 10,000,000). */
export function isValidSalary(salary: number): boolean {
  return salary > 0 && salary <= 10_000_000;
}

/** Return true if start is on or before end (YYYY-MM-DD strings). */
export function isDateRangeValid(start: string, end: string): boolean {
  return start <= end;
}

/** Return true if the start date is on or after the reference date (today by default). */
export function isStartDateValid(startDate: string, today?: string): boolean {
  const ref = today ?? todayIso();
  return startDate >= ref;
}

/**
 * Validate a leave request.
 * Returns null when valid, or an error message string when invalid.
 */
export function validateLeaveRequest(
  startDate: string,
  endDate: string,
  balance: number,
  days: number
): string | null {
  if (!isDateRangeValid(startDate, endDate)) return 'End date must be on or after start date.';
  if (days <= 0) return 'Days requested must be greater than zero.';
  if (days > balance) return 'Days requested exceed available balance.';
  return null;
}
