/** Format a 0–1 ratio as a percent string with sensible rounding. */
export function formatPercent(ratio: number, digits = 0): string {
  if (!Number.isFinite(ratio)) return '—';
  const pct = ratio * 100;
  return `${pct.toFixed(digits)}%`;
}
