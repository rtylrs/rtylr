// Pure helpers for discount calculation and validation.

/** Return the discounted amount after applying a percentage discount. */
export function applyPercent(amount: number, pct: number): number {
  if (!Number.isFinite(amount) || !Number.isFinite(pct)) return 0;
  const clamped = Math.min(Math.max(pct, 0), 100);
  return Math.max(0, amount * (1 - clamped / 100));
}

/** Return the discount amount from a percentage (not the discounted price). */
export function percentDiscountAmount(amount: number, pct: number): number {
  if (!Number.isFinite(amount) || !Number.isFinite(pct)) return 0;
  const clamped = Math.min(Math.max(pct, 0), 100);
  return Math.max(0, amount * (clamped / 100));
}

/** Return the discounted price after a flat dollar reduction. */
export function applyFlat(amount: number, flatDiscount: number): number {
  if (!Number.isFinite(amount) || !Number.isFinite(flatDiscount)) return Math.max(0, amount);
  return Math.max(0, amount - flatDiscount);
}

/** Return the effective discount percentage given the original amount and discount amount. */
export function effectiveDiscountPercent(original: number, discountAmount: number): number {
  if (!Number.isFinite(original) || original <= 0 || !Number.isFinite(discountAmount)) return 0;
  return Math.min(100, (discountAmount / original) * 100);
}

/** Return true if the discount percentage is within an acceptable range. */
export function isDiscountPermitted(pct: number, maxPct = 100): boolean {
  return Number.isFinite(pct) && pct >= 0 && pct <= maxPct;
}

/** Return true if the flat discount is within the order total. */
export function isFlatDiscountPermitted(flatDiscount: number, orderTotal: number): boolean {
  return Number.isFinite(flatDiscount) && flatDiscount >= 0 && flatDiscount <= orderTotal;
}

/** Stack two percentage discounts (apply second discount on the already-discounted price). */
export function stackPercentDiscounts(amount: number, pct1: number, pct2: number): number {
  return applyPercent(applyPercent(amount, pct1), pct2);
}
