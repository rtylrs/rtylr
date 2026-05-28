// Pure helpers for inventory and stock level calculations.

/** Returns true if stock is at or below the low-stock threshold. */
export function isLowStock(currentQty: number, threshold: number): boolean {
  if (!Number.isFinite(currentQty) || !Number.isFinite(threshold)) return false;
  return currentQty <= threshold;
}

/** Returns true if stock has reached zero or below. */
export function isOutOfStock(currentQty: number): boolean {
  return Number.isFinite(currentQty) && currentQty <= 0;
}

/** Clamp a stock adjustment to prevent going below zero.
 *  Returns the safe new quantity after applying the delta. */
export function applyStockDelta(currentQty: number, delta: number): number {
  if (!Number.isFinite(currentQty) || !Number.isFinite(delta)) return Math.max(0, currentQty || 0);
  return Math.max(0, currentQty + delta);
}

/** Calculate the quantity still needed to reach a target restock level. */
export function reorderQuantity(currentQty: number, targetQty: number): number {
  if (!Number.isFinite(currentQty) || !Number.isFinite(targetQty)) return 0;
  return Math.max(0, targetQty - currentQty);
}

/** Determine the stock status label for a given quantity and threshold. */
export function stockStatus(
  currentQty: number,
  threshold: number,
): 'out' | 'low' | 'ok' {
  if (isOutOfStock(currentQty)) return 'out';
  if (isLowStock(currentQty, threshold)) return 'low';
  return 'ok';
}

/** Format a stock quantity for display, showing "Out" when zero. */
export function formatStockQty(qty: number, unit = ''): string {
  if (!Number.isFinite(qty) || qty <= 0) return 'Out';
  return unit ? `${qty} ${unit}` : String(qty);
}
