const inFlight = new Map<string, Promise<unknown>>();

/**
 * Runs `fn` under an exclusive per-key guard. While a drain for `key` is in
 * flight, concurrent triggers (online event, reconnect broadcast, manual
 * flush) share the running promise instead of snapshotting the queue again
 * and POSTing duplicate entries.
 */
export function runExclusiveDrain<T>(key: string, fn: () => Promise<T>): Promise<T> {
  const existing = inFlight.get(key);
  if (existing) return existing as Promise<T>;
  const promise = fn().finally(() => {
    inFlight.delete(key);
  });
  inFlight.set(key, promise);
  return promise;
}
