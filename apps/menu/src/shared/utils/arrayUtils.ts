/**
 * Generic array utility helpers used across the menu app.
 */

/**
 * Returns a new array with duplicate values removed.
 * Uses strict equality for comparison.
 */
export function unique<T>(items: T[]): T[] {
  return [...new Set(items)];
}

/**
 * Splits an array into chunks of the given size.
 */
export function chunk<T>(items: T[], size: number): T[][] {
  if (size < 1) return [];
  const result: T[][] = [];
  for (let i = 0; i < items.length; i += size) {
    result.push(items.slice(i, i + size));
  }
  return result;
}

/**
 * Returns the first N items from an array.
 */
export function take<T>(items: T[], n: number): T[] {
  return items.slice(0, n);
}

/**
 * Partitions an array into two arrays: items that match the predicate and
 * items that do not.
 */
export function partition<T>(items: T[], predicate: (item: T) => boolean): [T[], T[]] {
  const pass: T[] = [];
  const fail: T[] = [];
  for (const item of items) {
    (predicate(item) ? pass : fail).push(item);
  }
  return [pass, fail];
}

/**
 * Flattens one level of nesting from an array of arrays.
 */
export function flatten<T>(items: T[][]): T[] {
  return ([] as T[]).concat(...items);
}

/**
 * Returns true if the array is empty or has no elements.
 */
export function isEmpty<T>(items: T[]): boolean {
  return items.length === 0;
}

/**
 * Returns a copy of the array with the element at the given index replaced.
 */
export function replaceAt<T>(items: T[], index: number, replacement: T): T[] {
  if (index < 0 || index >= items.length) return [...items];
  return [...items.slice(0, index), replacement, ...items.slice(index + 1)];
}

/**
 * Returns a copy of the array with the element at the given index removed.
 */
export function removeAt<T>(items: T[], index: number): T[] {
  if (index < 0 || index >= items.length) return [...items];
  return [...items.slice(0, index), ...items.slice(index + 1)];
}
