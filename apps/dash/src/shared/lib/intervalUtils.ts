/**
 * Time interval and scheduling utility helpers for the Dash app.
 * All functions are pure with no side-effects unless noted.
 */

export type IntervalUnit = 'ms' | 's' | 'min' | 'h' | 'd';

const UNIT_MS: Record<IntervalUnit, number> = {
  ms: 1,
  s: 1_000,
  min: 60_000,
  h: 3_600_000,
  d: 86_400_000,
};

/** Convert a value in the given unit to milliseconds. */
export function toMs(value: number, unit: IntervalUnit): number {
  return value * (UNIT_MS[unit] ?? 1);
}

/** Convert milliseconds to a human-readable duration string. */
export function msToHuman(ms: number): string {
  if (ms < 1000) return `${ms}ms`;
  if (ms < 60_000) return `${Math.round(ms / 1000)}s`;
  if (ms < 3_600_000) return `${Math.round(ms / 60_000)}min`;
  if (ms < 86_400_000) return `${Math.round(ms / 3_600_000)}h`;
  return `${Math.round(ms / 86_400_000)}d`;
}

/** Return the nearest safe polling interval in milliseconds (multiples of 5s). */
export function safePollingInterval(desiredMs: number): number {
  const step = 5_000;
  return Math.max(step, Math.round(desiredMs / step) * step);
}

/** Return true when the given timestamp (ms) is older than the threshold. */
export function isStale(timestampMs: number, thresholdMs: number, nowMs = Date.now()): boolean {
  return nowMs - timestampMs > thresholdMs;
}

/** Clamp a polling interval to the given min and max milliseconds. */
export function clampInterval(ms: number, minMs: number, maxMs: number): number {
  return Math.max(minMs, Math.min(maxMs, ms));
}

/** Return an exponential backoff delay for the given retry attempt (0-indexed). */
export function backoffDelay(attempt: number, baseMs = 1000, maxMs = 30_000): number {
  return clampInterval(baseMs * Math.pow(2, attempt), baseMs, maxMs);
}
