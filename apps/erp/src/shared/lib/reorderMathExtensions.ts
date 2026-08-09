/**
 * Extended reorder math utilities for the ERP application.
 * Supplements reorderMath.ts with additional supply chain calculations.
 */

/** Return the number of units to order to reach a target stock level. */
export function calculateOrderQuantity(
  currentStock: number,
  targetStock: number,
  minOrderQty = 1
): number {
  const needed = Math.max(0, targetStock - currentStock);
  if (needed === 0) return 0;
  return Math.max(minOrderQty, needed);
}

/** Return true when stock level has fallen to or below the reorder point. */
export function isAtReorderPoint(currentStock: number, reorderPoint: number): boolean {
  return currentStock <= reorderPoint;
}

/** Estimate the total order cost for a given quantity and unit cost. */
export function estimateOrderCost(quantity: number, unitCost: number): number {
  return quantity * unitCost;
}

/** Return the stock-to-sales ratio (days of inventory on hand). */
export function stockToSalesRatio(currentStock: number, avgDailySales: number): number {
  if (avgDailySales <= 0) return Infinity;
  return currentStock / avgDailySales;
}

/** Return the turnover rate: how many times stock is fully replaced per year. */
export function inventoryTurnoverRate(annualSales: number, averageStock: number): number {
  if (averageStock <= 0) return 0;
  return annualSales / averageStock;
}

/** Return the fill rate: percentage of demand that can be immediately satisfied. */
export function fillRate(available: number, demanded: number): number {
  if (demanded <= 0) return 100;
  return Math.min(100, (available / demanded) * 100);
}
