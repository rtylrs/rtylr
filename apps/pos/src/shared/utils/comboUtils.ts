// Pure helpers for combo display and pricing.

import type { Combo } from '../types/pos';

/** Return the total item count across all components in a combo. */
export function comboItemCount(combo: Combo): number {
  return combo.items.reduce((sum, i) => sum + i.qty, 0);
}

/** Return the price per component item (combo price divided by total item count). */
export function pricePerItem(combo: Combo): number {
  const count = comboItemCount(combo);
  if (count === 0 || !Number.isFinite(combo.price)) return 0;
  return combo.price / count;
}

/** Return true if a combo contains the given item id. */
export function comboContainsItem(combo: Combo, itemId: string): boolean {
  return combo.items.some((i) => i.itemId === itemId);
}

/** Return a comma-separated list of item names in a combo. */
export function comboItemNames(combo: Combo): string {
  return combo.items.map((i) => i.name ?? i.itemId).join(', ');
}

/** Return a display label with combo name and price. */
export function comboLabel(combo: Combo): string {
  return `${combo.name} - $${combo.price.toFixed(2)}`;
}

/** Sort combos by price ascending. */
export function sortCombosByPrice(combos: Combo[]): Combo[] {
  return [...combos].sort((a, b) => a.price - b.price);
}

/** Return combos that contain a specific item id. */
export function combosContainingItem(combos: Combo[], itemId: string): Combo[] {
  return combos.filter((c) => comboContainsItem(c, itemId));
}
