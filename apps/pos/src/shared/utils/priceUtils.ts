// Pure helpers for working with prices, taxes, and line item totals.

import type { CartItem } from '../stores/cartStore';

/** Calculate the total modifier cost for a single cart item. */
export function modifierTotal(item: Pick<CartItem, 'modifiers'>): number {
  return item.modifiers.reduce((sum, mod) => {
    const delta = Number.isFinite(mod.priceDelta) && mod.priceDelta > 0 ? mod.priceDelta : 0;
    return sum + delta;
  }, 0);
}

/** Calculate the effective unit price (base + modifiers) for a cart item. */
export function effectiveUnitPrice(item: Pick<CartItem, 'unitPrice' | 'modifiers'>): number {
  const base = Number.isFinite(item.unitPrice) && item.unitPrice >= 0 ? item.unitPrice : 0;
  return base + modifierTotal(item);
}

/** Calculate the gross line total (before discount) for a cart item. */
export function lineGross(item: Pick<CartItem, 'unitPrice' | 'quantity' | 'modifiers'>): number {
  return effectiveUnitPrice(item) * (Number.isFinite(item.quantity) && item.quantity > 0 ? item.quantity : 0);
}

/** Calculate the per-item discount amount for a cart item. */
export function itemDiscountAmount(
  item: Pick<CartItem, 'unitPrice' | 'quantity' | 'modifiers' | 'discountPct' | 'discountAmount'>,
): number {
  const gross = lineGross(item);
  if (item.discountPct > 0) {
    return Math.round(gross * (item.discountPct / 100) * 100) / 100;
  }
  if (item.discountAmount > 0) {
    return Math.min(gross, item.discountAmount);
  }
  return 0;
}

/** Calculate the net line total (after item-level discount) for a cart item. */
export function lineNet(
  item: Pick<CartItem, 'unitPrice' | 'quantity' | 'modifiers' | 'discountPct' | 'discountAmount'>,
): number {
  const gross = lineGross(item);
  const discount = itemDiscountAmount(item);
  return Math.max(0, Math.round((gross - discount) * 100) / 100);
}

/** Calculate the tax amount for a cart item after item-level discounts. */
export function lineTax(
  item: Pick<CartItem, 'unitPrice' | 'quantity' | 'modifiers' | 'discountPct' | 'discountAmount' | 'taxRate'>,
): number {
  const net = lineNet(item);
  const taxRate = Number.isFinite(item.taxRate) && item.taxRate >= 0 ? item.taxRate : 0;
  return Math.round(net * taxRate * 100) / 100;
}
