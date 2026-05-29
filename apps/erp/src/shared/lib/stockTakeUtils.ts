/**
 * Stock take helper functions used across the stock-takes feature.
 */
import type { StockTake } from '../services/erpApi';

/** Return true when a stock take is in a terminal state. */
export function isStockTakeComplete(stockTake: StockTake): boolean {
  return stockTake.status === 'Completed';
}

/** Return true when a stock take is currently active. */
export function isStockTakeActive(stockTake: StockTake): boolean {
  return stockTake.status === 'In Progress';
}

/** Return the percentage of items counted in a stock take, from 0 to 100. */
export function stockTakeProgressPct(stockTake: StockTake): number {
  if (stockTake.totalItems <= 0) return 0;
  return Math.min(100, Math.round((stockTake.itemsCounted / stockTake.totalItems) * 100));
}

/** Return true when the stock take has any variance (discrepancy). */
export function hasVariance(stockTake: StockTake): boolean {
  return stockTake.variance !== 0;
}

/** Return true when the stock take shows a negative variance (shrinkage). */
export function hasShrinkage(stockTake: StockTake): boolean {
  return stockTake.variance < 0;
}

/** Return a human-readable progress label such as "45 / 100 items". */
export function stockTakeProgressLabel(stockTake: StockTake): string {
  return `${stockTake.itemsCounted} / ${stockTake.totalItems} items`;
}

/** Sort stock takes by creation date, newest first. */
export function sortStockTakesByDateDesc(stockTakes: StockTake[]): StockTake[] {
  return [...stockTakes].sort(
    (a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
  );
}

/** Filter stock takes by status. */
export function filterStockTakesByStatus(
  stockTakes: StockTake[],
  status: StockTake['status']
): StockTake[] {
  return stockTakes.filter((s) => s.status === status);
}
