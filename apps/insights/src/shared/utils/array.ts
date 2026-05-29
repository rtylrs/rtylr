// Generic array utilities used across the insights feature modules.
// These are stateless helpers for grouping, deduplication, and
// ordering that don't belong in any specific domain module.

/**
 * groupBy groups an array of items by the return value of the key
 * function, producing a Map from key to array of items.
 */
export function groupBy<T, K>(items: T[], key: (item: T) => K): Map<K, T[]> {
  const map = new Map<K, T[]>();
  for (const item of items) {
    const k = key(item);
    const group = map.get(k) ?? [];
    group.push(item);
    map.set(k, group);
  }
  return map;
}

/**
 * uniqueBy returns an array with duplicates removed, comparing items
 * by the return value of the key function.
 */
export function uniqueBy<T, K>(items: T[], key: (item: T) => K): T[] {
  const seen = new Set<K>();
  return items.filter((item) => {
    const k = key(item);
    if (seen.has(k)) return false;
    seen.add(k);
    return true;
  });
}

/**
 * chunk splits an array into sequential sub-arrays of size n.
 * The last chunk may be smaller than n.
 */
export function chunk<T>(items: T[], n: number): T[][] {
  if (n <= 0) return [];
  const result: T[][] = [];
  for (let i = 0; i < items.length; i += n) {
    result.push(items.slice(i, i + n));
  }
  return result;
}

/**
 * sumBy sums the numeric values returned by the accessor function.
 */
export function sumBy<T>(items: T[], accessor: (item: T) => number): number {
  return items.reduce((total, item) => total + accessor(item), 0);
}

/**
 * maxBy returns the item with the highest accessor value, or undefined
 * when the array is empty.
 */
export function maxBy<T>(items: T[], accessor: (item: T) => number): T | undefined {
  if (items.length === 0) return undefined;
  return items.reduce((best, item) => (accessor(item) > accessor(best) ? item : best));
}

/**
 * minBy returns the item with the lowest accessor value, or undefined
 * when the array is empty.
 */
export function minBy<T>(items: T[], accessor: (item: T) => number): T | undefined {
  if (items.length === 0) return undefined;
  return items.reduce((best, item) => (accessor(item) < accessor(best) ? item : best));
}

/**
 * flatten flattens one level of nested arrays.
 */
export function flatten<T>(arrays: T[][]): T[] {
  return ([] as T[]).concat(...arrays);
}
