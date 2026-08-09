/** Returns unique elements from an array preserving original order. */
export function unique<T>(items: T[]): T[] {
  return [...new Set(items)];
}

/** Returns unique elements by a key function. */
export function uniqueBy<T>(items: T[], keyFn: (item: T) => unknown): T[] {
  const seen = new Set();
  return items.filter((item) => {
    const key = keyFn(item);
    if (seen.has(key)) return false;
    seen.add(key);
    return true;
  });
}

/** Chunks an array into groups of size n. */
export function chunk<T>(items: T[], size: number): T[][] {
  const chunks: T[][] = [];
  for (let i = 0; i < items.length; i += size) {
    chunks.push(items.slice(i, i + size));
  }
  return chunks;
}

/** Returns a flattened array one level deep. */
export function flatten<T>(items: T[][]): T[] {
  return items.reduce((acc, arr) => acc.concat(arr), []);
}

/** Groups items by a key function, returning a Map. */
export function groupBy<T>(items: T[], keyFn: (item: T) => string): Map<string, T[]> {
  const map = new Map<string, T[]>();
  for (const item of items) {
    const key = keyFn(item);
    if (!map.has(key)) map.set(key, []);
    map.get(key)!.push(item);
  }
  return map;
}

/** Returns items from the first array that are not in the second. */
export function difference<T>(a: T[], b: T[]): T[] {
  const setB = new Set(b);
  return a.filter((item) => !setB.has(item));
}

/** Returns items present in both arrays. */
export function intersection<T>(a: T[], b: T[]): T[] {
  const setB = new Set(b);
  return a.filter((item) => setB.has(item));
}

/** Returns the last element of an array, or undefined. */
export function last<T>(items: T[]): T | undefined {
  return items[items.length - 1];
}

/** Returns the first element of an array, or undefined. */
export function first<T>(items: T[]): T | undefined {
  return items[0];
}

/** Moves an item from one index to another, returning a new array. */
export function moveItem<T>(items: T[], fromIndex: number, toIndex: number): T[] {
  const result = [...items];
  const [removed] = result.splice(fromIndex, 1);
  result.splice(toIndex, 0, removed);
  return result;
}

/** Returns a shuffled copy of the array (Fisher-Yates). */
export function shuffle<T>(items: T[]): T[] {
  const result = [...items];
  for (let i = result.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1));
    [result[i], result[j]] = [result[j], result[i]];
  }
  return result;
}
