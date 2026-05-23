/** Tiny in-memory cache keyed by chat id. Used by features that don't need
 *  persistence - only fast access during a single page session. */
const cache = new Map<string, unknown>();

export function readContext<T>(key: string): T | undefined {
  return cache.get(key) as T | undefined;
}

export function writeContext<T>(key: string, value: T): void {
  cache.set(key, value);
}

export function clearContext(key?: string): void {
  if (key) cache.delete(key);
  else cache.clear();
}
