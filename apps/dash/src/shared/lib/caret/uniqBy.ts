/** Keep the first item for each value of `keyOf(item)`. */
export function uniqBy<T, K extends string | number>(items: T[], keyOf: (item: T) => K): T[] {
  const seen = new Set<K>();
  const out: T[] = [];
  for (const item of items) {
    const k = keyOf(item);
    if (seen.has(k)) continue;
    seen.add(k);
    out.push(item);
  }
  return out;
}
