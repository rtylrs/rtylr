/** Return the item that maximizes `selector`, or undefined for empty input. */
export function maxBy<T>(items: T[], selector: (item: T) => number): T | undefined {
  if (items.length === 0) return undefined;
  let best = items[0];
  let bestScore = selector(best);
  for (let i = 1; i < items.length; i++) {
    const s = selector(items[i]);
    if (s > bestScore) {
      best = items[i];
      bestScore = s;
    }
  }
  return best;
}
