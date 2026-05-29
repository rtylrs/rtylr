/**
 * Inventory valuation method helpers for the ERP app.
 * Supports FIFO, LIFO, and weighted average cost methods.
 * All functions are pure with no side-effects.
 */

export type ValuationMethod = 'FIFO' | 'LIFO' | 'WAC';

export interface CostLayer {
  quantity: number;
  unitCost: number;
  date: string;
}

/** Compute weighted average cost from a list of cost layers. */
export function weightedAverageCost(layers: CostLayer[]): number {
  const totalQty = layers.reduce((s, l) => s + l.quantity, 0);
  if (totalQty === 0) return 0;
  const totalCost = layers.reduce((s, l) => s + l.quantity * l.unitCost, 0);
  return totalCost / totalQty;
}

/** Compute FIFO cost for removing a quantity from layers (oldest first). */
export function fifoCost(layers: CostLayer[], quantityToRemove: number): number {
  const sorted = [...layers].sort((a, b) => a.date.localeCompare(b.date));
  let remaining = quantityToRemove;
  let totalCost = 0;

  for (const layer of sorted) {
    if (remaining <= 0) break;
    const used = Math.min(layer.quantity, remaining);
    totalCost += used * layer.unitCost;
    remaining -= used;
  }

  return totalCost;
}

/** Compute LIFO cost for removing a quantity from layers (newest first). */
export function lifoCost(layers: CostLayer[], quantityToRemove: number): number {
  const sorted = [...layers].sort((a, b) => b.date.localeCompare(a.date));
  let remaining = quantityToRemove;
  let totalCost = 0;

  for (const layer of sorted) {
    if (remaining <= 0) break;
    const used = Math.min(layer.quantity, remaining);
    totalCost += used * layer.unitCost;
    remaining -= used;
  }

  return totalCost;
}

/** Compute the total inventory value using the specified method. */
export function totalValuation(layers: CostLayer[], method: ValuationMethod): number {
  const totalQty = layers.reduce((s, l) => s + l.quantity, 0);
  switch (method) {
    case 'WAC':
      return weightedAverageCost(layers) * totalQty;
    case 'FIFO':
      return fifoCost(layers, totalQty);
    case 'LIFO':
      return lifoCost(layers, totalQty);
  }
}

/** Return the unit cost for a given method. */
export function unitCostForMethod(layers: CostLayer[], method: ValuationMethod): number {
  const totalQty = layers.reduce((s, l) => s + l.quantity, 0);
  if (totalQty === 0) return 0;
  return totalValuation(layers, method) / totalQty;
}

/** Compute gross profit given revenue and cost of goods sold. */
export function grossProfit(revenue: number, cogs: number): number {
  return revenue - cogs;
}

/** Compute gross profit margin as a percentage. */
export function grossProfitMarginPct(revenue: number, cogs: number): number {
  if (revenue === 0) return 0;
  return ((revenue - cogs) / revenue) * 100;
}
