// Pure helpers for menu item display and filtering.

import type { MenuItem } from '@voxire/voxie/pos/menu_item/v1/types';

/** Return true if a menu item is currently available. */
export function isAvailable(item: MenuItem): boolean {
  return item.isActive !== false;
}

/** Return true if a menu item is out of stock (stock is 0 or explicitly set to 0). */
export function isOutOfStock(item: MenuItem): boolean {
  return typeof item.stock === 'number' && item.stock <= 0;
}

/** Return true if a menu item has a low stock warning. */
export function isLowStock(item: MenuItem, threshold = 5): boolean {
  return typeof item.stock === 'number' && item.stock > 0 && item.stock <= threshold;
}

/** Return true if an item is available and in stock. */
export function isOrderable(item: MenuItem): boolean {
  return isAvailable(item) && !isOutOfStock(item);
}

/** Filter a list of menu items to only those that are orderable. */
export function orderableItems(items: MenuItem[]): MenuItem[] {
  return items.filter(isOrderable);
}

/** Group menu items by their category. */
export function itemsByCategory(items: MenuItem[]): Record<string, MenuItem[]> {
  const result: Record<string, MenuItem[]> = {};
  for (const item of items) {
    const key = item.category || 'Uncategorized';
    if (!result[key]) result[key] = [];
    result[key].push(item);
  }
  return result;
}

/** Return a sorted list of unique category names from a set of items. */
export function uniqueCategories(items: MenuItem[]): string[] {
  const cats = new Set(items.map((i) => i.category || 'Uncategorized'));
  return Array.from(cats).sort();
}

/** Return the price formatted as a display string (no currency symbol). */
export function priceLabel(item: MenuItem): string {
  return (item.salePrice ?? 0).toFixed(2);
}

/** Return the effective tax rate for an item (defaults to 0 if not set). */
export function itemTaxRate(item: MenuItem): number {
  const taxRate = item.taxRate ?? 0;
  return Number.isFinite(taxRate) ? taxRate : 0;
}
