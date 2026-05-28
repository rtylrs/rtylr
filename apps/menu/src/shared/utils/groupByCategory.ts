/**
 * Groups an array of items by a string key derived from each item.
 * Returns a Record mapping each key value to its array of items.
 */
export function groupBy<T>(items: T[], keyFn: (item: T) => string): Record<string, T[]> {
  const result: Record<string, T[]> = {};
  for (const item of items) {
    const key = keyFn(item);
    if (!result[key]) {
      result[key] = [];
    }
    result[key].push(item);
  }
  return result;
}

export interface Categorized {
  category: string;
}

/**
 * Groups menu items by their category field.
 * Items without a category are placed under the "uncategorized" key.
 */
export function groupByCategory<T extends Categorized>(items: T[]): Record<string, T[]> {
  return groupBy(items, (item) => item.category || 'uncategorized');
}

/**
 * Returns the unique, sorted list of category names present in the item array.
 */
export function categoryNames(items: Categorized[]): string[] {
  const names = new Set(items.map((i) => i.category || 'uncategorized'));
  return Array.from(names).sort();
}

/**
 * Returns items belonging to a specific category, case-insensitively.
 */
export function itemsForCategory<T extends Categorized>(items: T[], category: string): T[] {
  const lower = category.toLowerCase();
  return items.filter((i) => (i.category || 'uncategorized').toLowerCase() === lower);
}
