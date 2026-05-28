// Pure helpers for deriving summary info from a cart without the Zustand store.

import type { CartItem } from '../stores/cartStore';

/** Count the total number of items (sum of quantities) in the cart. */
export function totalItemCount(items: CartItem[]): number {
  return items.reduce((sum, item) => sum + item.quantity, 0);
}

/** Count the number of unique product lines (ignoring quantity). */
export function uniqueLineCount(items: CartItem[]): number {
  return items.length;
}

/** Return items whose KDS status matches the given status. */
export function itemsByStatus(
  items: CartItem[],
  status: CartItem['status'],
): CartItem[] {
  return items.filter((item) => item.status === status);
}

/** Return true if every item in the cart has been served or cancelled. */
export function allItemsSettled(items: CartItem[]): boolean {
  if (items.length === 0) return false;
  return items.every(
    (item) => item.status === 'served' || item.status === 'cancelled',
  );
}

/** Return items that have per-item discounts applied. */
export function discountedItems(items: CartItem[]): CartItem[] {
  return items.filter((item) => item.discountPct > 0 || item.discountAmount > 0);
}

/** Return a deduplicated list of category names from a list of items.
 *  (Assumes items carry a category-like field - not standard on CartItem,
 *  so we use item name prefix as a rough substitute via modifiers.) */
export function uniqueModifierGroups(items: CartItem[]): string[] {
  const groups = new Set<string>();
  for (const item of items) {
    for (const mod of item.modifiers) {
      if (mod.name) groups.add(mod.name);
    }
  }
  return Array.from(groups).sort();
}

/** Return the gross subtotal before any discounts (price * qty, no modifiers). */
export function grossSubtotal(items: CartItem[]): number {
  return items.reduce((sum, item) => sum + item.unitPrice * item.quantity, 0);
}
