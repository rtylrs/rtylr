/**
 * Generic array utility functions used across the ERP application.
 */

/** Group an array of objects by a key. */
export function groupBy<T>(items: T[], key: keyof T): Record<string, T[]> {
  return items.reduce<Record<string, T[]>>((acc, item) => {
    const k = String(item[key]);
    if (!acc[k]) acc[k] = [];
    acc[k].push(item);
    return acc;
  }, {});
}

/** Return unique values from an array. */
export function unique<T>(items: T[]): T[] {
  return [...new Set(items)];
}

/** Return unique items by a specific key. */
export function uniqueBy<T>(items: T[], key: keyof T): T[] {
  const seen = new Set<unknown>();
  return items.filter((item) => {
    const k = item[key];
    if (seen.has(k)) return false;
    seen.add(k);
    return true;
  });
}

/** Split an array into chunks of a given size. */
export function chunk<T>(items: T[], size: number): T[][] {
  if (size <= 0) return [];
  const result: T[][] = [];
  for (let i = 0; i < items.length; i += size) {
    result.push(items.slice(i, i + size));
  }
  return result;
}

/** Return the sum of a numeric field across an array of objects. */
export function sumBy<T>(items: T[], key: keyof T): number {
  return items.reduce((sum, item) => sum + Number(item[key]), 0);
}

/** Return the maximum value of a numeric field, or undefined for empty arrays. */
export function maxBy<T>(items: T[], key: keyof T): number | undefined {
  if (items.length === 0) return undefined;
  return Math.max(...items.map((item) => Number(item[key])));
}

/** Return the minimum value of a numeric field, or undefined for empty arrays. */
export function minBy<T>(items: T[], key: keyof T): number | undefined {
  if (items.length === 0) return undefined;
  return Math.min(...items.map((item) => Number(item[key])));
}

/** Return a sorted copy of an array by a key. */
export function sortBy<T>(items: T[], key: keyof T, dir: 'asc' | 'desc' = 'asc'): T[] {
  return [...items].sort((a, b) => {
    const av = a[key];
    const bv = b[key];
    if (av < bv) return dir === 'asc' ? -1 : 1;
    if (av > bv) return dir === 'asc' ? 1 : -1;
    return 0;
  });
}

/** Return the first item matching a predicate, or undefined. */
export function findFirst<T>(items: T[], predicate: (item: T) => boolean): T | undefined {
  return items.find(predicate);
}
