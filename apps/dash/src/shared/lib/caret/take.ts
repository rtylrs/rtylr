/** First N items (returns the full array when N >= length). */
export function take<T>(items: T[], n: number): T[] {
  if (n <= 0) return [];
  return items.slice(0, n);
}

/** Last N items. */
export function takeRight<T>(items: T[], n: number): T[] {
  if (n <= 0) return [];
  if (n >= items.length) return items.slice();
  return items.slice(items.length - n);
}
