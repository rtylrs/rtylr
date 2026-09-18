// Pure helpers for inventory item display and classification.

/** Inventory item shape used by these helpers. */
export interface InventoryItem {
  id: string;
  name: string;
  qty: number;
  unit: string;
  lowStockThreshold: number;
  category?: string;
  cost?: number;
}

/** Return true if the item is below its low-stock threshold. */
export function isLow(item: InventoryItem): boolean {
  return item.qty > 0 && item.qty <= item.lowStockThreshold;
}

/** Return true if the item has zero quantity. */
export function isEmpty(item: InventoryItem): boolean {
  return item.qty <= 0;
}

/** Return a status string for the item. */
export function itemStatus(item: InventoryItem): 'empty' | 'low' | 'ok' {
  if (isEmpty(item)) return 'empty';
  if (isLow(item)) return 'low';
  return 'ok';
}

/** Return the estimated total value of an item's stock. */
export function stockValue(item: InventoryItem): number {
  if (!Number.isFinite(item.cost) || !Number.isFinite(item.qty)) return 0;
  return Math.max(0, (item.cost ?? 0) * item.qty);
}

/** Return the quantity needed to reach a target level. */
export function reorderQty(item: InventoryItem, targetQty: number): number {
  if (!Number.isFinite(targetQty)) return 0;
  return Math.max(0, targetQty - item.qty);
}

/** Group inventory items by category. */
export function groupByCategory(items: InventoryItem[]): Record<string, InventoryItem[]> {
  const result: Record<string, InventoryItem[]> = {};
  for (const item of items) {
    const key = item.category ?? 'Uncategorized';
    if (!result[key]) result[key] = [];
    result[key].push(item);
  }
  return result;
}

/** Return only the items that need attention (empty or low). */
export function alertItems(items: InventoryItem[]): InventoryItem[] {
  return items.filter((i) => isLow(i) || isEmpty(i));
}

/** Return a display label with name, quantity, and unit. */
export function qtyLabel(item: InventoryItem): string {
  return `${item.name}: ${item.qty} ${item.unit}`;
}
