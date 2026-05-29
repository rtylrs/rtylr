/**
 * String masking helpers for sensitive data display in the Dash app.
 * All functions are pure with no side-effects.
 */

/** Mask all but the last N characters of a string with a fill character. */
export function maskEnd(value: string, visibleCount: number, fill = '*'): string {
  if (value.length <= visibleCount) return value;
  return fill.repeat(value.length - visibleCount) + value.slice(-visibleCount);
}

/** Mask all but the first N characters of a string. */
export function maskStart(value: string, visibleCount: number, fill = '*'): string {
  if (value.length <= visibleCount) return value;
  return value.slice(0, visibleCount) + fill.repeat(value.length - visibleCount);
}

/** Mask the middle of a string, keeping first and last N characters visible. */
export function maskMiddle(value: string, visibleEach: number, fill = '*'): string {
  if (value.length <= visibleEach * 2) return value;
  return (
    value.slice(0, visibleEach) +
    fill.repeat(value.length - visibleEach * 2) +
    value.slice(-visibleEach)
  );
}

/** Mask an email address, hiding most of the local part. */
export function maskEmail(email: string): string {
  const atIndex = email.indexOf('@');
  if (atIndex < 0) return email;
  const local = email.slice(0, atIndex);
  const domain = email.slice(atIndex);
  if (local.length <= 2) return local[0] + '*' + domain;
  return local[0] + '*'.repeat(local.length - 2) + local.slice(-1) + domain;
}

/** Mask a phone number, showing only the last 4 digits. */
export function maskPhone(phone: string): string {
  const digits = phone.replace(/\D/g, '');
  if (digits.length < 4) return '*'.repeat(digits.length);
  return '***-***-' + digits.slice(-4);
}

/** Mask a credit card number, showing only the last 4 digits. */
export function maskCard(card: string): string {
  const digits = card.replace(/\D/g, '');
  if (digits.length < 4) return '*'.repeat(digits.length);
  return '**** **** **** ' + digits.slice(-4);
}

/** Mask an API key, showing only the first 4 and last 4 characters. */
export function maskApiKey(key: string): string {
  if (key.length <= 8) return '*'.repeat(key.length);
  return key.slice(0, 4) + '*'.repeat(key.length - 8) + key.slice(-4);
}

/** Return true if a string looks like it is already masked (contains asterisks). */
export function isMasked(value: string): boolean {
  return value.includes('*');
}
