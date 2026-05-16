/** Group an array of records by the result of `keyOf(item)`. */
export function groupBy<T, K extends string>(items: T[], keyOf: (item: T) => K): Record<K, T[]> {
  const out: Record<K, T[]> = {} as Record<K, T[]>;
  for (const item of items) {
    const k = keyOf(item);
    if (!out[k]) out[k] = [];
    out[k].push(item);
  }
  return out;
}
