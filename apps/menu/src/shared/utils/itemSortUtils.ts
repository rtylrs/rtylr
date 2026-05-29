/**
 * Menu item sorting utilities for the Menu app.
 */

export type SortField = 'name' | 'price' | 'calories' | 'popular';
export type SortDirection = 'asc' | 'desc';

export interface SortableItem {
  name: string;
  price: number;
  calories?: number;
  orderCount?: number;
}

/** Compare two values for ascending sort (returns negative, 0, or positive). */
function compareAsc<T>(a: T, b: T): number {
  if (a < b) return -1;
  if (a > b) return 1;
  return 0;
}

/** Sort menu items by the given field and direction. */
export function sortItems<T extends SortableItem>(
  items: T[],
  field: SortField,
  direction: SortDirection = 'asc'
): T[] {
  const sorted = [...items].sort((a, b) => {
    switch (field) {
      case 'name':
        return compareAsc(a.name.toLowerCase(), b.name.toLowerCase());
      case 'price':
        return compareAsc(a.price, b.price);
      case 'calories':
        return compareAsc(a.calories ?? 0, b.calories ?? 0);
      case 'popular':
        return compareAsc(b.orderCount ?? 0, a.orderCount ?? 0);
      default:
        return 0;
    }
  });
  return direction === 'asc' ? sorted : sorted.reverse();
}

/** Return the top N items by order count. */
export function topItems<T extends SortableItem>(items: T[], n: number): T[] {
  return [...items].sort((a, b) => (b.orderCount ?? 0) - (a.orderCount ?? 0)).slice(0, n);
}

/** Return items within a price range (inclusive). */
export function filterByPriceRange<T extends SortableItem>(
  items: T[],
  min: number,
  max: number
): T[] {
  return items.filter((item) => item.price >= min && item.price <= max);
}
