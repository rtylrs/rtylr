/** Quick effort chips (minutes), pick a small win when energy is low. */
export const EFFORT_PRESETS = [15, 30, 60, 120, 240] as const;

export function formatEffortMinutes(m: number | undefined): string {
  if (m == null || m <= 0) return '-';
  if (m < 60) return `${m}m`;
  const h = m / 60;
  return Number.isInteger(h) ? `${h}h` : `${h.toFixed(1)}h`;
}
