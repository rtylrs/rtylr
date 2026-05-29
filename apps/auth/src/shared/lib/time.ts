/**
 * Returns the current time as a Unix timestamp in seconds.
 */
export function nowSeconds(): number {
  return Math.floor(Date.now() / 1000);
}

/**
 * Converts a number of days to seconds.
 */
export function daysToSeconds(days: number): number {
  return days * 24 * 60 * 60;
}

/**
 * Returns a future Unix timestamp (in seconds) that is `seconds` from now.
 */
export function futureTimestamp(seconds: number): number {
  return nowSeconds() + seconds;
}

/**
 * Returns true when the given Unix timestamp (in seconds) is in the past.
 */
export function isPast(timestampSeconds: number): boolean {
  return timestampSeconds < nowSeconds();
}

/**
 * Returns a human-readable relative time string.
 */
export function relativeTime(targetMs: number, nowMs = Date.now()): string {
  const diffMs = targetMs - nowMs;
  const diffSec = Math.round(diffMs / 1000);
  if (Math.abs(diffSec) < 5) return 'just now';
  const abs = Math.abs(diffSec);
  const past = diffSec < 0;
  let label: string;
  if (abs < 60) {
    label = `${abs} second${abs === 1 ? '' : 's'}`;
  } else if (abs < 3600) {
    const m = Math.round(abs / 60);
    label = `${m} minute${m === 1 ? '' : 's'}`;
  } else if (abs < 86400) {
    const h = Math.round(abs / 3600);
    label = `${h} hour${h === 1 ? '' : 's'}`;
  } else {
    const d = Math.round(abs / 86400);
    label = `${d} day${d === 1 ? '' : 's'}`;
  }
  return past ? `${label} ago` : `in ${label}`;
}
