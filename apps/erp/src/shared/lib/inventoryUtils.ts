/**
 * Inventory helper functions used across the ERP application.
 */
import type { InventoryReportItem } from '../services/erpApi';

/** Return the total inventory value across all items. */
export function totalInventoryValue(items: InventoryReportItem[]): number {
  return items.reduce((sum, item) => sum + item.value, 0);
}

/** Return the total current stock quantity across all inventory items. */
export function totalInventoryQuantity(items: InventoryReportItem[]): number {
  return items.reduce((sum, item) => sum + item.currentStock, 0);
}

/** Sort inventory items by value descending (highest value first). */
export function sortByValueDesc(items: InventoryReportItem[]): InventoryReportItem[] {
  return [...items].sort((a, b) => b.value - a.value);
}

/** Sort inventory items by stock ascending (lowest stock first). */
export function sortByStockAsc(items: InventoryReportItem[]): InventoryReportItem[] {
  return [...items].sort((a, b) => a.currentStock - b.currentStock);
}

/** Filter inventory items by SKU prefix (case-insensitive). */
export function filterBySkuPrefix(
  items: InventoryReportItem[],
  prefix: string
): InventoryReportItem[] {
  const p = prefix.toLowerCase();
  return items.filter((item) => item.sku.toLowerCase().startsWith(p));
}

/** Return the top N highest-value inventory items. */
export function topValueItems(items: InventoryReportItem[], n: number): InventoryReportItem[] {
  return sortByValueDesc(items).slice(0, n);
}

/** Return the percentage of total value represented by an item, from 0 to 100. */
export function valuePct(item: InventoryReportItem, total: number): number {
  if (total <= 0) return 0;
  return (item.value / total) * 100;
}

/** Return the average unit cost for an item (value / currentStock). */
export function averageUnitCost(item: InventoryReportItem): number {
  if (item.currentStock <= 0) return 0;
  return item.value / item.currentStock;
}
