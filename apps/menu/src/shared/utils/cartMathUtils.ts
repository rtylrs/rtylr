/**
 * Cart arithmetic helpers for the Menu app.
 * All functions are pure, working in cents to avoid floating-point drift.
 */

/** Convert a dollar amount to integer cents. */
export function toCents(dollars: number): number {
  return Math.round(dollars * 100);
}

/** Convert cents back to dollars. */
export function fromCents(cents: number): number {
  return cents / 100;
}

/** Compute the subtotal for a list of {price, quantity} line items (in dollars). */
export function subtotal(lines: Array<{ price: number; quantity: number }>): number {
  const cents = lines.reduce((acc, { price, quantity }) => acc + toCents(price) * quantity, 0);
  return fromCents(cents);
}

/** Apply a percentage discount to a dollar amount, returning the discounted amount. */
export function applyDiscountPercent(amount: number, pct: number): number {
  if (pct <= 0) return amount;
  if (pct >= 100) return 0;
  return fromCents(Math.round(toCents(amount) * (1 - pct / 100)));
}

/** Compute tax amount given a subtotal and a tax rate in percent. */
export function taxAmount(subtotalDollars: number, taxRatePct: number): number {
  return fromCents(Math.round(toCents(subtotalDollars) * (taxRatePct / 100)));
}

/** Compute the grand total: subtotal + tax - discount (all in dollars). */
export function grandTotal(
  subtotalDollars: number,
  taxRatePct: number,
  discountDollars = 0
): number {
  const subCents = toCents(subtotalDollars);
  const taxCents = Math.round(subCents * (taxRatePct / 100));
  const discCents = toCents(discountDollars);
  return fromCents(subCents + taxCents - discCents);
}

/** Count the total number of items across all cart lines. */
export function totalItemCount(lines: Array<{ quantity: number }>): number {
  return lines.reduce((acc, { quantity }) => acc + quantity, 0);
}
