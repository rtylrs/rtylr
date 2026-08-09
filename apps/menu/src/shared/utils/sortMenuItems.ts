export interface SortableItem {
  name: string;
  salePrice?: number;
  isActive?: boolean;
}

/**
 * Sorts menu items alphabetically by name (A to Z).
 */
export function sortByNameAsc<T extends SortableItem>(items: T[]): T[] {
  return [...items].sort((a, b) => a.name.localeCompare(b.name));
}

/**
 * Sorts menu items alphabetically by name (Z to A).
 */
export function sortByNameDesc<T extends SortableItem>(items: T[]): T[] {
  return [...items].sort((a, b) => b.name.localeCompare(a.name));
}

/**
 * Sorts menu items by price, lowest first.
 * Items without a salePrice are placed last.
 */
export function sortByPriceAsc<T extends SortableItem>(items: T[]): T[] {
  return [...items].sort((a, b) => {
    const pa = a.salePrice ?? Infinity;
    const pb = b.salePrice ?? Infinity;
    return pa - pb;
  });
}

/**
 * Sorts menu items by price, highest first.
 * Items without a salePrice are placed last.
 */
export function sortByPriceDesc<T extends SortableItem>(items: T[]): T[] {
  return [...items].sort((a, b) => {
    const pa = a.salePrice ?? -Infinity;
    const pb = b.salePrice ?? -Infinity;
    return pb - pa;
  });
}

/**
 * Filters out inactive items, returning only active ones.
 */
export function filterActiveItems<T extends SortableItem>(items: T[]): T[] {
  return items.filter((i) => i.isActive !== false);
}
