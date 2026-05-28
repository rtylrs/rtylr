// Pure helpers for order number formatting and parsing.

/** Pad an order number to a fixed width with leading zeros. */
export function padOrderNumber(num: number, width = 4): string {
  if (!Number.isFinite(num) || num < 0) return '0'.repeat(width);
  return String(Math.floor(num)).padStart(width, '0');
}

/** Format an order number with a prefix (e.g. "#0042"). */
export function formatOrderNumber(num: number, prefix = '#', width = 4): string {
  return `${prefix}${padOrderNumber(num, width)}`;
}

/** Parse the numeric value out of a formatted order number string. */
export function parseOrderNumber(formatted: string): number | null {
  const digits = formatted.replace(/\D/g, '');
  if (!digits) return null;
  const num = parseInt(digits, 10);
  return Number.isFinite(num) ? num : null;
}

/** Return true if two formatted order numbers refer to the same order. */
export function isSameOrder(a: string, b: string): boolean {
  const na = parseOrderNumber(a);
  const nb = parseOrderNumber(b);
  if (na === null || nb === null) return false;
  return na === nb;
}

/** Generate an order number from a base timestamp (seconds since epoch). */
export function orderNumberFromTimestamp(epochSeconds: number, modulus = 10000): number {
  if (!Number.isFinite(epochSeconds) || epochSeconds < 0) return 0;
  return Math.floor(epochSeconds) % modulus;
}

/** Return true if a formatted order number string matches the expected pattern. */
export function isValidOrderNumberString(input: string): boolean {
  return /^#?\d{1,6}$/.test(input.trim());
}
