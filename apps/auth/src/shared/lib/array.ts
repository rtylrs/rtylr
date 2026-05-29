/**
 * Generic array helpers used across the auth app.
 */

/**
 * Returns a new array with duplicate values removed (preserves first occurrence).
 */
export function unique<T>(items: T[]): T[] {
  return [...new Set(items)];
}

/**
 * Splits an array into chunks of the given size.
 */
export function chunk<T>(items: T[], size: number): T[][] {
  if (size <= 0) return [];
  const result: T[][] = [];
  for (let i = 0; i < items.length; i += size) {
    result.push(items.slice(i, i + size));
  }
  return result;
}

/**
 * Returns the last element of an array, or undefined if empty.
 */
export function last<T>(items: T[]): T | undefined {
  return items[items.length - 1];
}

/**
 * Returns the first element of an array, or undefined if empty.
 */
export function first<T>(items: T[]): T | undefined {
  return items[0];
}

/**
 * Groups an array of items by the value returned from the key function.
 */
export function groupBy<T>(items: T[], key: (item: T) => string): Record<string, T[]> {
  return items.reduce<Record<string, T[]>>((acc, item) => {
    const k = key(item);
    if (!acc[k]) acc[k] = [];
    acc[k].push(item);
    return acc;
  }, {});
}

/**
 * Returns true when two arrays contain the same elements (order-insensitive).
 */
export function sameElements<T>(a: T[], b: T[]): boolean {
  if (a.length !== b.length) return false;
  const setA = new Set(a);
  return b.every((item) => setA.has(item));
}

/**
 * Returns elements in a that are not in b.
 */
export function difference<T>(a: T[], b: T[]): T[] {
  const setB = new Set(b);
  return a.filter((item) => !setB.has(item));
}
