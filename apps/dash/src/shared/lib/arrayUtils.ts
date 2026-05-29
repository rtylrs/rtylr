/**
 * Array utility helpers for the Dash app.
 * All functions are pure with no side-effects.
 */

/** Split an array into chunks of the given size. */
export function chunk<T>(arr: T[], size: number): T[][] {
  if (size <= 0) return [];
  const result: T[][] = [];
  for (let i = 0; i < arr.length; i += size) {
    result.push(arr.slice(i, i + size));
  }
  return result;
}

/** Return the first n items of an array. */
export function take<T>(arr: T[], n: number): T[] {
  return arr.slice(0, Math.max(0, n));
}

/** Return the last n items of an array. */
export function takeLast<T>(arr: T[], n: number): T[] {
  return arr.slice(Math.max(0, arr.length - n));
}

/** Sort an array by a key function returning a string or number. Does not mutate. */
export function sortBy<T>(arr: T[], key: (item: T) => number | string): T[] {
  return [...arr].sort((a, b) => {
    const ka = key(a);
    const kb = key(b);
    if (ka < kb) return -1;
    if (ka > kb) return 1;
    return 0;
  });
}

/** Partition an array into [matches, non-matches]. */
export function partition<T>(arr: T[], predicate: (item: T) => boolean): [T[], T[]] {
  const matches: T[] = [];
  const rest: T[] = [];
  for (const item of arr) {
    if (predicate(item)) matches.push(item);
    else rest.push(item);
  }
  return [matches, rest];
}

/** Flatten a 2D array into a 1D array. */
export function flatten<T>(arr: T[][]): T[] {
  return ([] as T[]).concat(...arr);
}

/** Return unique values from an array (preserves order). */
export function unique<T>(arr: T[]): T[] {
  return [...new Set(arr)];
}

/** Zip two arrays together into an array of pairs. */
export function zip<A, B>(a: A[], b: B[]): [A, B][] {
  const len = Math.min(a.length, b.length);
  const result: [A, B][] = [];
  for (let i = 0; i < len; i++) {
    result.push([a[i] as A, b[i] as B]);
  }
  return result;
}
