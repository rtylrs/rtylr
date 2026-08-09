/**
 * Phone number formatting and validation helpers for the menu app.
 */

/**
 * Strips all non-digit characters from a phone string.
 */
export function digitsOnly(phone: string): string {
  return phone.replace(/\D/g, '');
}

/**
 * Returns true when the input contains at least 7 and at most 15 digits
 * (ITU-T E.164 range).
 */
export function isValidPhoneLength(phone: string): boolean {
  const digits = digitsOnly(phone);
  return digits.length >= 7 && digits.length <= 15;
}

/**
 * Formats a 10-digit US phone number as (XXX) XXX-XXXX.
 * Returns the original string if it is not exactly 10 digits.
 */
export function formatUsPhone(phone: string): string {
  const digits = digitsOnly(phone);
  if (digits.length !== 10) return phone;
  return `(${digits.slice(0, 3)}) ${digits.slice(3, 6)}-${digits.slice(6)}`;
}

/**
 * Normalises a phone string to E.164 format by prepending "+" and removing
 * non-digit characters. Assumes country code is already included in the input.
 */
export function toE164(phone: string): string {
  const digits = digitsOnly(phone);
  return `+${digits}`;
}

/**
 * Returns a masked display version of a phone number, showing only the last
 * four digits. E.g. "555-867-5309" becomes "***-***-5309".
 */
export function maskPhone(phone: string): string {
  const digits = digitsOnly(phone);
  if (digits.length < 4) return '****';
  const visible = digits.slice(-4);
  const hidden = '*'.repeat(digits.length - 4);
  return `${hidden}${visible}`;
}
