/** Sum the numeric result of `selector(item)` across `items`. */
export function sumBy<T>(items: T[], selector: (item: T) => number): number {
  let total = 0;
  for (const item of items) total += selector(item);
  return total;
}
