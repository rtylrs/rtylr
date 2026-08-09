export interface SearchableItem {
  name: string;
  sku?: string;
  [key: string]: unknown;
}

/** Return true if the item matches the given search query. */
export function itemMatchesQuery(item: SearchableItem, query: string): boolean {
  const q = query.trim().toLowerCase();
  if (!q) return true;
  if (item.name.toLowerCase().includes(q)) return true;
  if (item.sku && item.sku.toLowerCase().includes(q)) return true;
  return false;
}

/** Filter a list of menu items by a search query string. */
export function searchMenuItems<T extends SearchableItem>(items: T[], query: string): T[] {
  const q = query.trim().toLowerCase();
  if (!q) return items;
  return items.filter((item) => itemMatchesQuery(item, q));
}
