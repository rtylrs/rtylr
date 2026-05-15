const UNITS = ['B', 'KB', 'MB', 'GB'];

/** Format a byte count as a human-readable size string. */
export function formatBytes(n: number): string {
  if (!Number.isFinite(n) || n < 0) return '0 B';
  let i = 0;
  let value = n;
  while (value >= 1024 && i < UNITS.length - 1) {
    value /= 1024;
    i++;
  }
  return `${value.toFixed(value < 10 && i > 0 ? 1 : 0)} ${UNITS[i]}`;
}
