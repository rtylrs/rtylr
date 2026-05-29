/**
 * Safe localStorage wrapper for the auth app.
 */

/**
 * Returns true when localStorage is available in the current environment.
 */
export function hasLocalStorage(): boolean {
  try {
    return typeof window !== 'undefined' && typeof window.localStorage !== 'undefined';
  } catch {
    return false;
  }
}

/**
 * Reads a raw string value from localStorage.
 * Returns null when the key does not exist or storage is unavailable.
 */
export function storageGet(key: string): string | null {
  if (!hasLocalStorage()) return null;
  try {
    return window.localStorage.getItem(key);
  } catch {
    return null;
  }
}

/**
 * Writes a string value to localStorage.
 */
export function storageSet(key: string, value: string): void {
  if (!hasLocalStorage()) return;
  try {
    window.localStorage.setItem(key, value);
  } catch {
    // ignore quota errors
  }
}

/**
 * Removes a key from localStorage.
 */
export function storageRemove(key: string): void {
  if (!hasLocalStorage()) return;
  try {
    window.localStorage.removeItem(key);
  } catch {
    // ignore
  }
}

/**
 * Reads and JSON-parses a value from localStorage.
 * Returns the fallback when the key is missing or the value is not valid JSON.
 */
export function storageGetJson<T>(key: string, fallback: T): T {
  const raw = storageGet(key);
  if (raw === null) return fallback;
  try {
    return JSON.parse(raw) as T;
  } catch {
    return fallback;
  }
}

/**
 * JSON-serialises a value and writes it to localStorage.
 */
export function storageSetJson<T>(key: string, value: T): void {
  try {
    storageSet(key, JSON.stringify(value));
  } catch {
    // ignore serialisation errors
  }
}
