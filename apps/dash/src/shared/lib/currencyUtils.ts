/**
 * Currency and money formatting helpers for the Dash app.
 * All functions are pure with no side-effects.
 */

/** Format a number as a currency string using Intl.NumberFormat. */
export function formatCurrency(amount: number, currency = 'USD', locale = 'en-US'): string {
  const safeAmount = Number.isFinite(amount) ? amount : 0;
  return new Intl.NumberFormat(locale, {
    style: 'currency',
    currency,
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(safeAmount);
}

/** Round a monetary value to 2 decimal places (avoids floating point drift). */
export function roundMoney(amount: number): number {
  return Math.round(amount * 100) / 100;
}

/** Compute a discount amount from a price and percentage. */
export function discountAmount(price: number, discountPct: number): number {
  return roundMoney((price * Math.min(100, Math.max(0, discountPct))) / 100);
}

/** Apply a percentage tax to an amount and return the total. */
export function applyTax(amount: number, taxPct: number): number {
  return roundMoney(amount * (1 + taxPct / 100));
}

/** Return the tax portion of a tax-inclusive amount. */
export function extractTax(totalWithTax: number, taxPct: number): number {
  if (taxPct === 0) return 0;
  return roundMoney(totalWithTax - totalWithTax / (1 + taxPct / 100));
}

/** Compute the subtotal from a list of line items. */
export function lineItemsSubtotal(items: { quantity: number; unitPrice: number }[]): number {
  return roundMoney(items.reduce((sum, item) => sum + item.quantity * item.unitPrice, 0));
}

/** Return true if the amount is a positive finite number. */
export function isValidAmount(amount: unknown): amount is number {
  return typeof amount === 'number' && Number.isFinite(amount) && amount > 0;
}
