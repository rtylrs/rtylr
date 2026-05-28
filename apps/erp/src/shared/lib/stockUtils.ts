/**
 * Stock-level helper functions used across inventory, products, and alerts.
 */

/** Return true when the current quantity is strictly below the low-stock threshold. */
export function isLowStock(quantity: number, threshold: number): boolean {
  return quantity < threshold;
}

/** Return true when stock is completely depleted. */
export function isOutOfStock(quantity: number): boolean {
  return quantity <= 0;
}

/** Return a stock status label matching the ERP product status enum. */
export function stockStatusLabel(
  quantity: number,
  threshold: number
): 'In Stock' | 'Low Stock' | 'Out of Stock' {
  if (isOutOfStock(quantity)) return 'Out of Stock';
  if (isLowStock(quantity, threshold)) return 'Low Stock';
  return 'In Stock';
}

/** Compute the stock deficit: how many units are needed to reach the target level. */
export function stockDeficit(current: number, target: number): number {
  return Math.max(0, target - current);
}

/** Compute the stock surplus: how many units exceed the target level. */
export function stockSurplus(current: number, target: number): number {
  return Math.max(0, current - target);
}

/** Percentage of capacity used, clamped to [0, 100]. */
export function capacityUsedPct(used: number, capacity: number): number {
  if (capacity <= 0) return 0;
  return Math.min(100, Math.round((used / capacity) * 100));
}

/** Return the inventory value for a set of units at a given cost. */
export function inventoryValue(quantity: number, costPrice: number): number {
  return quantity * costPrice;
}
