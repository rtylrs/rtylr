/**
 * Price list and pricing rule helpers for the ERP app.
 * All functions are pure with no side-effects.
 */

export type PriceListType = 'standard' | 'wholesale' | 'promotional' | 'member';

export interface PriceListEntry {
  productId: number;
  sku: string;
  basePrice: number;
  listPrice: number;
  currency: string;
  minQuantity?: number;
}

export interface PricingRule {
  type: 'percentage' | 'fixed' | 'multiplier';
  value: number;
}

/** Apply a pricing rule to a base price. */
export function applyPricingRule(basePrice: number, rule: PricingRule): number {
  switch (rule.type) {
    case 'percentage':
      return basePrice * (1 - rule.value / 100);
    case 'fixed':
      return basePrice - rule.value;
    case 'multiplier':
      return basePrice * rule.value;
  }
}

/** Compute the discount amount from a base price and list price. */
export function discountAmount(basePrice: number, listPrice: number): number {
  return Math.max(0, basePrice - listPrice);
}

/** Compute the discount percentage from a base price and list price. */
export function discountPct(basePrice: number, listPrice: number): number {
  if (basePrice === 0) return 0;
  return Math.max(0, ((basePrice - listPrice) / basePrice) * 100);
}

/** Return true if a price list entry qualifies for a given quantity. */
export function qualifiesForEntry(entry: PriceListEntry, quantity: number): boolean {
  return quantity >= (entry.minQuantity ?? 1);
}

/** Find the best (lowest) price for a product across multiple entries for a given quantity. */
export function bestPrice(entries: PriceListEntry[], productId: number, quantity = 1): number {
  const qualifying = entries
    .filter((e) => e.productId === productId && qualifiesForEntry(e, quantity))
    .map((e) => e.listPrice);
  if (qualifying.length === 0) return 0;
  return Math.min(...qualifying);
}

/** Group entries by product id. */
export function groupByProduct(entries: PriceListEntry[]): Map<number, PriceListEntry[]> {
  const map = new Map<number, PriceListEntry[]>();
  for (const entry of entries) {
    const existing = map.get(entry.productId) ?? [];
    map.set(entry.productId, [...existing, entry]);
  }
  return map;
}

/** Format a price with currency symbol. */
export function formatListPrice(price: number, currency = 'USD'): string {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency,
    minimumFractionDigits: 2,
  }).format(price);
}

/** Return entries sorted by list price ascending. */
export function sortByPrice(entries: PriceListEntry[]): PriceListEntry[] {
  return [...entries].sort((a, b) => a.listPrice - b.listPrice);
}
