// Pure array utility helpers used across the POS codebase.

/** Group an array of items by a key function. */
export function groupBy<T>(items: T[], keyFn: (item: T) => string): Record<string, T[]> {
  const result: Record<string, T[]> = {};
  for (const item of items) {
    const key = keyFn(item);
    if (!result[key]) result[key] = [];
    result[key].push(item);
  }
  return result;
}

/** Sum a numeric field across an array of objects. */
export function sumBy<T>(items: T[], valueFn: (item: T) => number): number {
  return items.reduce((sum, item) => sum + (Number.isFinite(valueFn(item)) ? valueFn(item) : 0), 0);
}

/** Return the item with the maximum value for a given numeric field. */
export function maxBy<T>(items: T[], valueFn: (item: T) => number): T | undefined {
  if (items.length === 0) return undefined;
  return items.reduce((best, item) => (valueFn(item) > valueFn(best) ? item : best));
}

/** Return the item with the minimum value for a given numeric field. */
export function minBy<T>(items: T[], valueFn: (item: T) => number): T | undefined {
  if (items.length === 0) return undefined;
  return items.reduce((best, item) => (valueFn(item) < valueFn(best) ? item : best));
}

/** Chunk an array into sub-arrays of a given size. */
export function chunk<T>(items: T[], size: number): T[][] {
  if (size <= 0) return [];
  const result: T[][] = [];
  for (let i = 0; i < items.length; i += size) {
    result.push(items.slice(i, i + size));
  }
  return result;
}

/** Return a new array with the item at the given index moved to a new position. */
export function moveItem<T>(items: T[], fromIndex: number, toIndex: number): T[] {
  if (fromIndex === toIndex) return [...items];
  if (fromIndex < 0 || fromIndex >= items.length) return [...items];
  if (toIndex < 0 || toIndex >= items.length) return [...items];
  const result = [...items];
  const [removed] = result.splice(fromIndex, 1);
  result.splice(toIndex, 0, removed);
  return result;
}

/** Return a sorted copy of an array using a numeric value function (ascending). */
export function sortByValue<T>(items: T[], valueFn: (item: T) => number): T[] {
  return [...items].sort((a, b) => valueFn(a) - valueFn(b));
}
