/**
 * Currency helpers for the menu app.
 * Wraps Intl.NumberFormat to keep formatting consistent.
 */

/**
 * Returns the currency symbol for a given ISO 4217 currency code.
 * Falls back to the code itself when the runtime cannot resolve it.
 */
export function currencySymbol(code: string, locale = 'en-US'): string {
  try {
    const parts = new Intl.NumberFormat(locale, {
      style: 'currency',
      currency: code,
      minimumFractionDigits: 0,
      maximumFractionDigits: 0,
    }).formatToParts(0);
    const sym = parts.find((p) => p.type === 'currency');
    return sym ? sym.value : code;
  } catch {
    return code;
  }
}

/**
 * Rounds a number to the given number of decimal places.
 */
export function roundTo(value: number, decimals = 2): number {
  const factor = Math.pow(10, decimals);
  return Math.round(value * factor) / factor;
}

/**
 * Converts cents (integer) to a decimal dollar amount.
 */
export function centsToDollars(cents: number): number {
  return roundTo(cents / 100, 2);
}

/**
 * Converts a decimal dollar amount to cents (integer).
 */
export function dollarsToCents(dollars: number): number {
  return Math.round(dollars * 100);
}

/**
 * Applies a percentage discount to a price and returns the discounted value.
 * The percent parameter should be a value between 0 and 100.
 */
export function applyDiscount(price: number, percentOff: number): number {
  if (percentOff <= 0) return price;
  if (percentOff >= 100) return 0;
  return roundTo(price * (1 - percentOff / 100), 2);
}

/**
 * Calculates the sales tax amount for a given price and tax rate.
 * taxRate should be a percentage value, e.g. 8.5 for 8.5%.
 */
export function calcTax(price: number, taxRate: number): number {
  return roundTo(price * (taxRate / 100), 2);
}
