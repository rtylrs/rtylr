/** Now-string in ISO-8601 for chat persistence. Centralized so tests can
 *  patch via vi.useFakeTimers without sprinkling Date.now everywhere. */
export function nowIso(): string {
  return new Date().toISOString();
}

/** Now in epoch ms for diagnostics. */
export function nowEpoch(): number {
  return Date.now();
}
