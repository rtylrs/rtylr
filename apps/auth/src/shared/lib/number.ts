/**
 * Number formatting and parsing helpers for the auth app.
 */

/**
 * Formats a number using the given locale (default en-US).
 */
export function formatNumber(value: number, locale = 'en-US'): string {
  return new Intl.NumberFormat(locale).format(value);
}

/**
 * Formats a number as a currency string.
 */
export function formatCurrency(value: number, currency = 'USD', locale = 'en-US'): string {
  return new Intl.NumberFormat(locale, { style: 'currency', currency }).format(value);
}

/**
 * Formats a number as a percentage string (e.g. 0.42 -> "42%").
 */
export function formatPercent(value: number, locale = 'en-US'): string {
  return new Intl.NumberFormat(locale, { style: 'percent' }).format(value);
}

/**
 * Parses a string containing digits (and optional decimal separator) into a number.
 * Returns NaN when no numeric content is found.
 */
export function parseNumeric(value: string): number {
  const cleaned = value.replace(/[^0-9.-]/g, '');
  return cleaned === '' ? NaN : Number(cleaned);
}

/**
 * Returns true when the value is a finite number (not NaN or Infinity).
 */
export function isValidNumber(value: unknown): value is number {
  return typeof value === 'number' && isFinite(value);
}
