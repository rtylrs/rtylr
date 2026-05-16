/** Generate a list of half-hour buckets across the past N hours for analytics charts. */
export function pastHourBuckets(hours = 24): { from: number; to: number }[] {
  const now = Date.now();
  const halfHour = 30 * 60_000;
  const out: { from: number; to: number }[] = [];
  for (let i = hours * 2 - 1; i >= 0; i--) {
    const to = now - i * halfHour;
    out.push({ from: to - halfHour, to });
  }
  return out;
}
