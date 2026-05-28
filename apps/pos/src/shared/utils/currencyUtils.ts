// Currency formatting and parsing helpers for POS display.

/** Parse a currency string like "$12.50" or "12.50" to a number. */
export function parseCurrencyString(input: string): number | null {
  const cleaned = input.replace(/[$,\s]/g, '').trim();
  if (!cleaned) return null;
  const num = Number(cleaned);
  if (!Number.isFinite(num)) return null;
  return num;
}

/** Format a number as a currency string with the given symbol. */
export function formatWithSymbol(amount: number, symbol = '$'): string {
  if (!Number.isFinite(amount)) return `${symbol}0.00`;
  const abs = Math.abs(amount);
  const formatted = `${symbol}${abs.toFixed(2)}`;
  return amount < 0 ? `-${formatted}` : formatted;
}

/** Return the absolute value for display (e.g., for showing refund amounts positively). */
export function absAmount(amount: number): number {
  if (!Number.isFinite(amount)) return 0;
  return Math.abs(amount);
}

/** Convert cents (integer) to dollars. */
export function centsToDollars(cents: number): number {
  if (!Number.isFinite(cents)) return 0;
  return Math.round(cents) / 100;
}

/** Convert dollars to cents (integer). */
export function dollarsToCents(dollars: number): number {
  if (!Number.isFinite(dollars)) return 0;
  return Math.round(dollars * 100);
}

/** Check if two currency amounts are effectively equal (within floating point tolerance). */
export function amountsEqual(a: number, b: number, tolerance = 0.001): boolean {
  if (!Number.isFinite(a) || !Number.isFinite(b)) return false;
  return Math.abs(a - b) <= tolerance;
}
