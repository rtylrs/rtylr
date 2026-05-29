/**
 * Pricing utility functions used across the ERP application.
 */

/** Return the gross margin percentage given cost and sale prices. */
export function grossMarginPct(costPrice: number, salePrice: number): number {
  if (salePrice <= 0) return 0;
  return ((salePrice - costPrice) / salePrice) * 100;
}

/** Return the markup percentage over cost price. */
export function markupPct(costPrice: number, salePrice: number): number {
  if (costPrice <= 0) return 0;
  return ((salePrice - costPrice) / costPrice) * 100;
}

/** Return a sale price derived from a cost price and desired markup percentage. */
export function saleFromMarkup(costPrice: number, markupPercent: number): number {
  return costPrice * (1 + markupPercent / 100);
}

/** Apply a multiplier to a base price (e.g. 1.2 for 20% uplift). */
export function applyMultiplier(basePrice: number, multiplier: number): number {
  if (multiplier < 0) return 0;
  return basePrice * multiplier;
}

/** Apply a percentage discount to a price. */
export function applyDiscount(price: number, discountPct: number): number {
  if (discountPct <= 0) return price;
  if (discountPct >= 100) return 0;
  return price * (1 - discountPct / 100);
}

/** Return the profit per unit. */
export function profitPerUnit(costPrice: number, salePrice: number): number {
  return salePrice - costPrice;
}

/** Round a price to 2 decimal places. */
export function roundPrice(price: number): number {
  return Math.round((price + Number.EPSILON) * 100) / 100;
}

/** Return true when a price is a valid positive number. */
export function isValidPrice(price: number): boolean {
  return isFinite(price) && price >= 0;
}
