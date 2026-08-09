/**
 * Price display and formatting helpers for the Menu app.
 */

/** Format a number as a price string with currency symbol. */
export function formatPrice(amount: number, currency = 'USD', locale = 'en-US'): string {
  return new Intl.NumberFormat(locale, {
    style: 'currency',
    currency,
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(amount);
}

/** Return the price difference between original and discounted values. */
export function priceDiff(original: number, discounted: number): number {
  return Math.max(0, original - discounted);
}

/** Return the discount percentage between original and discounted values (0-100). */
export function discountPct(original: number, discounted: number): number {
  if (original <= 0) return 0;
  return Math.round(((original - discounted) / original) * 100);
}

/** Return true if a price has an active discount. */
export function hasDiscount(original: number, discounted: number): boolean {
  return discounted < original;
}

/** Return the effective unit price after dividing by quantity. */
export function unitPrice(total: number, quantity: number): number {
  if (quantity <= 0) return 0;
  return total / quantity;
}

/** Return a compact display string like "$12.50 / item". */
export function perItemLabel(price: number, currency = 'USD', locale = 'en-US'): string {
  return `${formatPrice(price, currency, locale)} / item`;
}

/** Round a price to the nearest cent. */
export function roundPrice(amount: number): number {
  // Nudge by epsilon so exact half-cents (e.g. 1.005) round up despite
  // floating point representation error.
  return Math.round((amount + Number.EPSILON) * 100) / 100;
}
